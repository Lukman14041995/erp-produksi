package catalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranji/clothing-erp/internal/db"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ---- Product types (Jenis Pesanan) ----

const productTypeColumns = `id, code, name, sort_order, is_active, created_at, updated_at`

func scanProductType(row pgx.Row) (ProductType, error) {
	var p ProductType
	err := row.Scan(&p.ID, &p.Code, &p.Name, &p.SortOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) CreateProductType(ctx context.Context, in UpsertProductTypeInput) (ProductType, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO product_types (code, name, sort_order)
		VALUES ($1,$2,$3)
		RETURNING `+productTypeColumns, in.Code, in.Name, in.SortOrder)
	return scanProductType(row)
}

func (r *Repository) UpdateProductType(ctx context.Context, id uuid.UUID, in UpsertProductTypeInput) (ProductType, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE product_types SET name=$2, sort_order=$3, is_active=$4
		WHERE id=$1
		RETURNING `+productTypeColumns, id, in.Name, in.SortOrder, isActive)
	return scanProductType(row)
}

func (r *Repository) GetProductTypeByID(ctx context.Context, id uuid.UUID) (ProductType, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+productTypeColumns+` FROM product_types WHERE id=$1`, id)
	return scanProductType(row)
}

func (r *Repository) ListProductTypes(ctx context.Context, activeOnly bool) ([]ProductType, error) {
	q := `SELECT ` + productTypeColumns + ` FROM product_types`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProductType{}
	for rows.Next() {
		p, err := scanProductType(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---- Fabrics ----

const fabricColumns = `id, product_type_id, code, material_id, name, sales_price, image_url, sort_order, is_active, created_at, updated_at`

func scanFabric(row pgx.Row) (Fabric, error) {
	var f Fabric
	err := row.Scan(&f.ID, &f.ProductTypeID, &f.Code, &f.MaterialID, &f.Name, &f.SalesPrice, &f.ImageURL, &f.SortOrder, &f.IsActive, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func (r *Repository) CreateFabric(ctx context.Context, in UpsertFabricInput) (Fabric, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO fabrics (product_type_id, code, material_id, name, sales_price, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+fabricColumns, in.ProductTypeID, in.Code, in.MaterialID, in.Name, in.SalesPrice, in.SortOrder)
	return scanFabric(row)
}

func (r *Repository) UpdateFabric(ctx context.Context, id uuid.UUID, in UpsertFabricInput) (Fabric, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE fabrics SET product_type_id=$2, material_id=$3, name=$4, sales_price=$5, sort_order=$6, is_active=$7
		WHERE id=$1
		RETURNING `+fabricColumns, id, in.ProductTypeID, in.MaterialID, in.Name, in.SalesPrice, in.SortOrder, isActive)
	return scanFabric(row)
}

func (r *Repository) SetFabricImage(ctx context.Context, id uuid.UUID, imageURL string) (Fabric, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE fabrics SET image_url=$2 WHERE id=$1
		RETURNING `+fabricColumns, id, imageURL)
	return scanFabric(row)
}

func (r *Repository) GetFabricByID(ctx context.Context, id uuid.UUID) (Fabric, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+fabricColumns+` FROM fabrics WHERE id=$1`, id)
	return scanFabric(row)
}

func (r *Repository) ListFabrics(ctx context.Context, activeOnly bool) ([]Fabric, error) {
	q := `SELECT ` + fabricColumns + ` FROM fabrics`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Fabric{}
	for rows.Next() {
		f, err := scanFabric(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// ---- Garment variants (Model potongan) ----

const variantColumns = `id, product_type_id, code, name, price_addon, icon_url, sort_order, is_active, created_at, updated_at`

func scanVariant(row pgx.Row) (GarmentVariant, error) {
	var v GarmentVariant
	err := row.Scan(&v.ID, &v.ProductTypeID, &v.Code, &v.Name, &v.PriceAddon, &v.IconURL, &v.SortOrder, &v.IsActive, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}

func (r *Repository) CreateVariant(ctx context.Context, in UpsertVariantInput) (GarmentVariant, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO garment_variants (product_type_id, code, name, price_addon, sort_order)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING `+variantColumns, in.ProductTypeID, in.Code, in.Name, in.PriceAddon, in.SortOrder)
	return scanVariant(row)
}

func (r *Repository) UpdateVariant(ctx context.Context, id uuid.UUID, in UpsertVariantInput) (GarmentVariant, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE garment_variants SET product_type_id=$2, name=$3, price_addon=$4, sort_order=$5, is_active=$6
		WHERE id=$1
		RETURNING `+variantColumns, id, in.ProductTypeID, in.Name, in.PriceAddon, in.SortOrder, isActive)
	return scanVariant(row)
}

func (r *Repository) SetVariantIcon(ctx context.Context, id uuid.UUID, iconURL string) (GarmentVariant, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE garment_variants SET icon_url=$2 WHERE id=$1
		RETURNING `+variantColumns, id, iconURL)
	return scanVariant(row)
}

func (r *Repository) GetVariantByID(ctx context.Context, id uuid.UUID) (GarmentVariant, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+variantColumns+` FROM garment_variants WHERE id=$1`, id)
	return scanVariant(row)
}

func (r *Repository) ListVariants(ctx context.Context, activeOnly bool) ([]GarmentVariant, error) {
	q := `SELECT ` + variantColumns + ` FROM garment_variants`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GarmentVariant{}
	for rows.Next() {
		v, err := scanVariant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ---- Garment sizes ----

const garmentSizeColumns = `id, size_code, size_multiplier, sort_order, is_active, created_at, updated_at`

func scanGarmentSize(row pgx.Row) (GarmentSize, error) {
	var s GarmentSize
	err := row.Scan(&s.ID, &s.SizeCode, &s.SizeMultiplier, &s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) CreateGarmentSize(ctx context.Context, in UpsertSizeInput) (GarmentSize, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO garment_sizes (size_code, size_multiplier, sort_order)
		VALUES ($1,$2,$3)
		RETURNING `+garmentSizeColumns, in.SizeCode, in.SizeMultiplier, in.SortOrder)
	return scanGarmentSize(row)
}

func (r *Repository) UpdateGarmentSize(ctx context.Context, id uuid.UUID, in UpsertSizeInput) (GarmentSize, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE garment_sizes SET size_code=$2, size_multiplier=$3, sort_order=$4, is_active=$5
		WHERE id=$1
		RETURNING `+garmentSizeColumns, id, in.SizeCode, in.SizeMultiplier, in.SortOrder, isActive)
	return scanGarmentSize(row)
}

func (r *Repository) GetGarmentSizeByID(ctx context.Context, id uuid.UUID) (GarmentSize, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+garmentSizeColumns+` FROM garment_sizes WHERE id=$1`, id)
	return scanGarmentSize(row)
}

func (r *Repository) ListGarmentSizes(ctx context.Context, activeOnly bool) ([]GarmentSize, error) {
	q := `SELECT ` + garmentSizeColumns + ` FROM garment_sizes`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GarmentSize{}
	for rows.Next() {
		s, err := scanGarmentSize(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ---- Inks (Tinta) ----

const inkColumns = `id, product_type_id, material_id, code, name, price_addon, image_url, sort_order, is_active, created_at, updated_at`

func scanInk(row pgx.Row) (Ink, error) {
	var i Ink
	err := row.Scan(&i.ID, &i.ProductTypeID, &i.MaterialID, &i.Code, &i.Name, &i.PriceAddon, &i.ImageURL, &i.SortOrder, &i.IsActive, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (r *Repository) CreateInk(ctx context.Context, in UpsertInkInput) (Ink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO inks (product_type_id, material_id, code, name, price_addon, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+inkColumns, in.ProductTypeID, in.MaterialID, in.Code, in.Name, in.PriceAddon, in.SortOrder)
	return scanInk(row)
}

func (r *Repository) UpdateInk(ctx context.Context, id uuid.UUID, in UpsertInkInput) (Ink, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE inks SET product_type_id=$2, material_id=$3, name=$4, price_addon=$5, sort_order=$6, is_active=$7
		WHERE id=$1
		RETURNING `+inkColumns, id, in.ProductTypeID, in.MaterialID, in.Name, in.PriceAddon, in.SortOrder, isActive)
	return scanInk(row)
}

func (r *Repository) SetInkImage(ctx context.Context, id uuid.UUID, imageURL string) (Ink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `
		UPDATE inks SET image_url=$2 WHERE id=$1
		RETURNING `+inkColumns, id, imageURL)
	return scanInk(row)
}

func (r *Repository) GetInkByID(ctx context.Context, id uuid.UUID) (Ink, error) {
	row := db.Q(ctx, r.pool).QueryRow(ctx, `SELECT `+inkColumns+` FROM inks WHERE id=$1`, id)
	return scanInk(row)
}

func (r *Repository) ListInks(ctx context.Context, activeOnly bool) ([]Ink, error) {
	q := `SELECT ` + inkColumns + ` FROM inks`
	if activeOnly {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY sort_order`
	rows, err := db.Q(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ink{}
	for rows.Next() {
		i, err := scanInk(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
