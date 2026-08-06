package qr

// QROptions defines configuration parameters for QR code generation.
type QROptions struct {
	Size            int    `json:"size"`             // Dimensions in pixels (default 256)
	Format          string `json:"format"`           // "png" or "svg"
	ErrorCorrection string `json:"error_correction"` // "L", "M", "Q", "H"
	FGColor         string `json:"fg_color,omitempty"`
	BGColor         string `json:"bg_color,omitempty"`
}
