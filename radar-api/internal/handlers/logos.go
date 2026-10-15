package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ArminDashti/radar-api/internal/models"
	"github.com/gin-gonic/gin"
	xdraw "golang.org/x/image/draw"
)

const logoSize = 64

func (s *Server) UploadHostLogo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host id"})
		return
	}

	var exists bool
	if err := s.DB.QueryRowContext(requestContext(c), `SELECT EXISTS(SELECT 1 FROM endpoints WHERE id = ?)`, id).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo file is required"})
		return
	}
	if file.Size > 4<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo must be 4MB or smaller"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read logo"})
		return
	}
	defer src.Close()

	payload, err := io.ReadAll(io.LimitReader(src, 4<<20+1))
	if err != nil || len(payload) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read logo"})
		return
	}
	if len(payload) > 4<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logo must be 4MB or smaller"})
		return
	}

	img, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image format"})
		return
	}

	processed := processLogo(img)
	if err := os.MkdirAll(s.LogoDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not prepare logo storage"})
		return
	}

	filename := fmt.Sprintf("%d.png", id)
	path := filepath.Join(s.LogoDir, filename)
	out, err := os.Create(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store logo"})
		return
	}
	defer out.Close()
	if err := png.Encode(out, processed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not encode logo"})
		return
	}

	logoPath := "/api/logos/" + filename
	var host models.Endpoint
	err = s.DB.QueryRowContext(requestContext(c), `
		UPDATE endpoints SET logo_icon = ? WHERE id = ?
		RETURNING id, name, host, http_enabled, icmp_enabled, probe_id, active, COALESCE(logo_icon, ''), created_at`,
		logoPath, id,
	).Scan(&host.ID, &host.Name, &host.Host, &host.HTTPEnabled, &host.ICMPEnabled, &host.ProbeID, &host.Active, &host.LogoIcon, &host.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update host logo"})
		return
	}
	c.JSON(http.StatusOK, host)
}

func (s *Server) ServeLogo(c *gin.Context) {
	name := filepath.Base(c.Param("filename"))
	if name == "." || name == string(filepath.Separator) || strings.Contains(name, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid logo name"})
		return
	}
	if !strings.HasSuffix(strings.ToLower(name), ".png") {
		c.JSON(http.StatusNotFound, gin.H{"error": "logo not found"})
		return
	}
	path := filepath.Join(s.LogoDir, name)
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "logo not found"})
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(path)
}

func processLogo(src image.Image) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, logoSize, logoSize))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	makeBackgroundTransparent(dst)
	return dst
}

func makeBackgroundTransparent(img *image.NRGBA) {
	bounds := img.Bounds()
	if bounds.Dx() < 2 || bounds.Dy() < 2 {
		return
	}
	samples := []color.NRGBA{
		img.NRGBAAt(bounds.Min.X, bounds.Min.Y),
		img.NRGBAAt(bounds.Max.X-1, bounds.Min.Y),
		img.NRGBAAt(bounds.Min.X, bounds.Max.Y-1),
		img.NRGBAAt(bounds.Max.X-1, bounds.Max.Y-1),
	}
	bg := averageOpaque(samples)
	if bg.A == 0 {
		return
	}
	const threshold = 48.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := img.NRGBAAt(x, y)
			if px.A == 0 {
				continue
			}
			if colorDistance(px, bg) <= threshold {
				px.A = 0
				img.SetNRGBA(x, y, px)
			}
		}
	}
}

func averageOpaque(samples []color.NRGBA) color.NRGBA {
	var r, g, b, n int
	for _, sample := range samples {
		if sample.A < 32 {
			continue
		}
		r += int(sample.R)
		g += int(sample.G)
		b += int(sample.B)
		n++
	}
	if n == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n), A: 255}
}

func colorDistance(a, b color.NRGBA) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}
