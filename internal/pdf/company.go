package pdf

// Company is the static letterhead profile used on every generated document.
// The ERP has no company-settings module yet, so this is a fixed profile
// rather than something loaded from the database.
var Company = struct {
	Name              string
	Tagline           string
	Address           string
	Phone             string
	Email             string
	TaxID             string
	BankName          string
	BankAccountName   string
	BankAccountNumber string
}{
	Name:              "PT Ranji Garment Industries",
	Tagline:           "Custom Jersey & Apparel Manufacturer",
	Address:           "Jl. Industri Tekstil No. 88, Bandung, Jawa Barat 40213",
	Phone:             "+62 22 1234 5678",
	Email:             "info@ranjigarment.co.id",
	TaxID:             "NPWP: 01.234.567.8-901.000",
	BankName:          "Bank Mandiri",
	BankAccountName:   "PT Ranji Garment Industries",
	BankAccountNumber: "123-00-4567890-1",
}
