package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/ranji/clothing-erp/internal/accounting"
	"github.com/ranji/clothing-erp/internal/auth"
	"github.com/ranji/clothing-erp/internal/coa"
	"github.com/ranji/clothing-erp/internal/config"
	"github.com/ranji/clothing-erp/internal/customer"
	"github.com/ranji/clothing-erp/internal/db"
	"github.com/ranji/clothing-erp/internal/finance"
	"github.com/ranji/clothing-erp/internal/inventory"
	"github.com/ranji/clothing-erp/internal/material"
	appmw "github.com/ranji/clothing-erp/internal/middleware"
	"github.com/ranji/clothing-erp/internal/migrate"
	"github.com/ranji/clothing-erp/internal/product"
	"github.com/ranji/clothing-erp/internal/production"
	"github.com/ranji/clothing-erp/internal/purchasing"
	"github.com/ranji/clothing-erp/internal/reporting"
	"github.com/ranji/clothing-erp/internal/response"
	"github.com/ranji/clothing-erp/internal/sales"
	"github.com/ranji/clothing-erp/internal/supplier"
)

func main() {
	cfg := config.Load()

	// Migrations run before anything else connects, so a fresh database (or
	// one with pending schema changes) is ready by the time the pool below
	// starts serving requests. Safe under multiple concurrently starting
	// replicas -- golang-migrate serializes via a Postgres advisory lock.
	if cfg.MigrateOnStartup {
		if err := migrate.Run(cfg.DatabaseURL); err != nil {
			log.Fatalf("run database migrations: %v", err)
		}
		log.Println("database migrations up to date")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	// ---- Wire repositories ----
	authRepo := auth.NewRepository(pool)
	coaRepo := coa.NewRepository(pool)
	journalRepo := accounting.NewRepository(pool)
	periodRepo := accounting.NewPeriodRepository(pool)
	customerRepo := customer.NewRepository(pool)
	supplierRepo := supplier.NewRepository(pool)
	productRepo := product.NewRepository(pool)
	materialRepo := material.NewRepository(pool)
	invRepo := inventory.NewRepository(pool)
	salesRepo := sales.NewRepository(pool)
	financeRepo := finance.NewRepository(pool)
	productionRepo := production.NewRepository(pool)
	purchasingRepo := purchasing.NewRepository(pool)
	reportingRepo := reporting.NewRepository(pool)

	// ---- Wire services (dependency order matters: accounting first, then
	// domains that post through it, then domains that depend on those) ----
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	coaSvc := coa.NewService(pool, coaRepo)
	accSvc := accounting.NewService(pool, journalRepo, coaRepo, periodRepo)
	customerSvc := customer.NewService(pool, customerRepo)
	supplierSvc := supplier.NewService(pool, supplierRepo)
	productSvc := product.NewService(pool, productRepo)
	materialSvc := material.NewService(pool, materialRepo)
	invSvc := inventory.NewService(pool, invRepo)
	salesSvc := sales.NewService(pool, salesRepo, accSvc, invSvc)
	purchasingSvc := purchasing.NewService(pool, purchasingRepo, coaRepo, accSvc, invSvc)
	financeSvc := finance.NewService(pool, financeRepo, accSvc, coaRepo, salesSvc, purchasingSvc)
	productionSvc := production.NewService(pool, productionRepo, productRepo, accSvc, invSvc, salesSvc)
	reportingSvc := reporting.NewService(reportingRepo)

	// ---- Wire handlers ----
	authHandler := auth.NewHandler(authSvc)
	coaHandler := coa.NewHandler(coaSvc)
	accHandler := accounting.NewHandler(accSvc, periodRepo)
	customerHandler := customer.NewHandler(customerSvc)
	supplierHandler := supplier.NewHandler(supplierSvc)
	productHandler := product.NewHandler(productSvc)
	materialHandler := material.NewHandler(materialSvc)
	invHandler := inventory.NewHandler(invSvc, invRepo, accSvc)
	salesHandler := sales.NewHandler(salesSvc, customerSvc, productSvc)
	purchasingHandler := purchasing.NewHandler(purchasingSvc, supplierSvc, materialSvc)
	financeHandler := finance.NewHandler(financeSvc)
	productionHandler := production.NewHandler(productionSvc, productSvc, materialSvc, salesSvc)
	reportingHandler := reporting.NewHandler(reportingSvc)

	e := echo.New()
	e.HTTPErrorHandler = appmw.HTTPErrorHandler
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// /health is a liveness+readiness probe in one: it pings the database so
	// an orchestrator (docker-compose healthcheck, k8s readiness probe)
	// doesn't route traffic to a replica that's up but can't reach Postgres.
	e.GET("/health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return response.Fail(c, http.StatusServiceUnavailable, "database unreachable", err.Error())
		}
		return response.OK(c, http.StatusOK, "ok", map[string]string{"status": "healthy"})
	})

	// ---- Public auth routes (no JWT required) ----
	public := e.Group("/api/v1")
	authHandler.RegisterPublic(public)

	// ---- Protected routes: every domain below requires a valid access
	// token. RBAC is layered on top per domain via RequireRole, at
	// domain-level granularity (all of a domain's routes share one role
	// gate) rather than per HTTP verb. ----
	requireAuth := auth.RequireAuth(authSvc)
	protected := e.Group("/api/v1", requireAuth, appmw.Actor())

	authHandler.RegisterProtected(protected)

	// Master data and reporting: readable/writable by any authenticated role.
	openToAll := protected.Group("")
	coaHandler.Register(openToAll)
	customerHandler.Register(openToAll)
	supplierHandler.Register(openToAll)
	productHandler.Register(openToAll)
	materialHandler.Register(openToAll)
	reportingHandler.Register(openToAll)
	accHandler.RegisterReadOnly(openToAll)

	// Sales: order/invoice lifecycle owned by SALES (and ADMIN).
	salesGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleSales))
	salesHandler.Register(salesGroup)

	// Production: production orders, BOM, HPP owned by PRODUCTION (and ADMIN).
	productionGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleProduction))
	productionHandler.Register(productionGroup)

	// Inventory: touched by production (issue/receive) and finance (adjustments/valuation).
	inventoryGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleProduction, auth.RoleFinance))
	invHandler.Register(inventoryGroup)

	// Purchasing (AP subledger): finance owns payables, production requests materials.
	purchasingGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleFinance, auth.RoleProduction))
	purchasingHandler.Register(purchasingGroup)

	// Finance: payments and expenses owned by FINANCE (and ADMIN).
	financeGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleFinance))
	financeHandler.Register(financeGroup)

	// Accounting: journals and period close owned by ACCOUNTING (and ADMIN).
	accountingGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin, auth.RoleAccounting))
	accHandler.Register(accountingGroup)

	// User management: ADMIN only.
	adminGroup := protected.Group("", auth.RequireRole(auth.RoleAdmin))
	authHandler.RegisterAdmin(adminGroup)

	go func() {
		log.Printf("listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received %s, draining in-flight requests...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown did not complete cleanly: %v", err)
	} else {
		log.Println("http server stopped")
	}
	// deferred pool.Close() below runs after this point, closing the DB pool
	// only once all in-flight requests have finished or the timeout elapsed.
	log.Println("closing database connection pool")
}
