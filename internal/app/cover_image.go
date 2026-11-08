// ABOUTME: Resolves, bounds, optionally upscales, and prepares script cover images.
// ABOUTME: Produces renderer-neutral physical geometry and temporary treated image variants.
package app

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
)

const (
	coverSourceByteLimit     int64  = 100 * 1024 * 1024
	coverSourceAxisLimit     uint32 = 32768
	coverSourcePixelLimit    uint64 = 100_000_000
	coverBackdropJPEGQuality        = 85
)

type coverImagePlan struct {
	Config                                                    config.CoverImage
	SourcePath, OriginalSourcePath, BlurredPath, BackdropPath string
	PixelWidth, PixelHeight                                   uint32
	RenderedWidthMM, RenderedHeightMM                         float64
	PageWidthMM, PageHeightMM                                 float64
}

type preparedCoverVariants struct {
	ImagePath, BlurredPath, BackdropPath string
}

func validateCoverSourceLimits(bytes int64, width, height uint32) error {
	if bytes < 1 || bytes > coverSourceByteLimit {
		return fmt.Errorf("folio.cover-image.path size %d bytes exceeds the permitted range 1 through %d", bytes, coverSourceByteLimit)
	}
	if width == 0 || height == 0 {
		return fmt.Errorf("folio.cover-image.path dimension must be at least 1 pixel; got %dx%d", width, height)
	}
	if width > coverSourceAxisLimit || height > coverSourceAxisLimit {
		return fmt.Errorf("folio.cover-image.path dimension %dx%d exceeds the per-axis limit %d", width, height, coverSourceAxisLimit)
	}
	pixels := uint64(width) * uint64(height)
	if pixels > coverSourcePixelLimit {
		return fmt.Errorf("folio.cover-image.path decoded pixel count %d exceeds the limit %d", pixels, coverSourcePixelLimit)
	}
	return nil
}

func planCoverImage(cfg config.Config, sourceDocument string) (*coverImagePlan, error) {
	cover, enabled, err := cfg.CoverImage()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}
	if sourceDocument == "" || sourceDocument == "-" {
		return nil, fmt.Errorf("folio.cover-image.path cannot be resolved for relative input without a source document path")
	}
	path := cover.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(sourceDocument), path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving folio.cover-image.path: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("folio.cover-image.path %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("folio.cover-image.path %q must be a regular PNG or JPEG file", path)
	}
	if info.Size() < 1 || info.Size() > coverSourceByteLimit {
		return nil, fmt.Errorf("folio.cover-image.path size %d bytes exceeds the permitted range 1 through %d", info.Size(), coverSourceByteLimit)
	}
	width, height, format, err := coverDecodeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("folio.cover-image.path %q is not a supported PNG or JPEG: %w", path, err)
	}
	if format != "png" && format != "jpeg" {
		return nil, fmt.Errorf("folio.cover-image.path %q has unsupported image format %q; expected PNG or JPEG", path, format)
	}
	if err := validateCoverSourceLimits(info.Size(), width, height); err != nil {
		return nil, err
	}
	plan := &coverImagePlan{Config: cover, SourcePath: path, OriginalSourcePath: path, PixelWidth: width, PixelHeight: height}
	plan.PageWidthMM, plan.PageHeightMM = coverPageDimensionsMM(cfg.String("folio.page", "a4"))
	if cover.Mode == "inset" {
		size, err := millimetres(cover.Size)
		if err != nil {
			return nil, err
		}
		scale := size / float64(maxUint32(width, height))
		plan.RenderedWidthMM = float64(width) * scale
		plan.RenderedHeightMM = float64(height) * scale
	} else {
		plan.RenderedWidthMM, plan.RenderedHeightMM = plan.PageWidthMM, plan.PageHeightMM
	}
	return plan, nil
}

func prepareCoverForRender(cfg config.Config, sourceDocument string, stderr io.Writer) (*coverImagePlan, func(), error) {
	plan, err := planCoverImage(cfg, sourceDocument)
	if err != nil || plan == nil {
		return plan, func() {}, err
	}
	tempDir, err := os.MkdirTemp("", "folio-cover-*")
	if err != nil {
		return nil, func() {}, fmt.Errorf("creating folio.cover-image temporary directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }
	selected := plan.SourcePath
	if plan.Config.Superscale != "" {
		output, ok, superscaleErr := runCoverSuperscale(plan.Config.Superscale, selected, tempDir, stderr)
		if superscaleErr != nil {
			cleanup()
			return nil, func() {}, superscaleErr
		}
		if ok {
			selected = output
			width, height, _, decodeErr := coverDecodeConfig(output)
			if decodeErr != nil {
				cleanup()
				return nil, func() {}, fmt.Errorf("decoding selected folio.cover-image.superscale output: %w", decodeErr)
			}
			plan.PixelWidth, plan.PixelHeight = width, height
		}
	}
	plan.SourcePath = selected
	variants, err := prepareCoverVariants(plan, tempDir)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	plan.SourcePath, plan.BlurredPath, plan.BackdropPath = variants.ImagePath, variants.BlurredPath, variants.BackdropPath
	return plan, cleanup, nil
}

func runCoverSuperscale(value, input, tempDir string, stderr io.Writer) (string, bool, error) {
	executable, err := exec.LookPath("superscale")
	if err != nil {
		coverSuperscaleWarning(stderr, "executable was not found in inherited PATH")
		return "", false, nil
	}
	outputDir := filepath.Join(tempDir, "superscale-output")
	if err := os.Mkdir(outputDir, 0o700); err != nil {
		coverSuperscaleWarning(stderr, fmt.Sprintf("temporary output could not be created: %v", err))
		return "", false, nil
	}
	args := []string{"--no-face-enhance", "-o", outputDir}
	if value != "default" {
		args = append(args, "--scale", strings.TrimSuffix(value, "x"))
	}
	args = append(args, input)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout, cmd.Stderr = stderr, stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", false, fmt.Errorf("folio.cover-image.superscale was interrupted: %w", ctx.Err())
		}
		coverSuperscaleWarning(stderr, fmt.Sprintf("execution failed: %v", err))
		return "", false, nil
	}
	output, err := usableSuperscaleOutput(outputDir)
	if err != nil {
		coverSuperscaleWarning(stderr, err.Error())
		return "", false, nil
	}
	return output, true, nil
}

func coverSuperscaleWarning(stderr io.Writer, reason string) {
	if stderr == nil {
		stderr = io.Discard
	}
	fmt.Fprintf(stderr, "Warning: folio.cover-image.superscale %s; using the original image.\n", reason)
}

func usableSuperscaleOutput(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("output directory could not be read: %v", err)
	}
	var candidates []string
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		width, height, format, err := coverDecodeConfig(path)
		if err == nil && width > 0 && height > 0 && (format == "png" || format == "jpeg") {
			candidates = append(candidates, path)
		}
	}
	if len(candidates) != 1 {
		return "", fmt.Errorf("produced %d usable PNG or JPEG outputs instead of exactly one", len(candidates))
	}
	return candidates[0], nil
}

func prepareCoverVariants(plan *coverImagePlan, dir string) (preparedCoverVariants, error) {
	decoded, format, err := decodeCoverImage(plan.SourcePath)
	if err != nil {
		return preparedCoverVariants{}, err
	}
	if format != "png" && format != "jpeg" {
		return preparedCoverVariants{}, fmt.Errorf("folio.cover-image.path has unsupported image format %q", format)
	}
	treated := treatCoverImage(decoded, plan.Config.ImageTreatment, plan.Config.ImageOpacity)
	treatedPath := filepath.Join(dir, "prepared.png")
	if err := writeCoverPNG(treatedPath, treated); err != nil {
		return preparedCoverVariants{}, err
	}
	blurredPath := treatedPath
	blurMM, err := millimetres(plan.Config.OverlayBlur)
	if err != nil {
		return preparedCoverVariants{}, err
	}
	if blurMM > 0 {
		pixelsPerMM := 0.0
		if plan.RenderedWidthMM > 0 {
			pixelsPerMM = float64(plan.PixelWidth) / plan.RenderedWidthMM
		}
		if plan.RenderedHeightMM > 0 && float64(plan.PixelHeight)/plan.RenderedHeightMM > pixelsPerMM {
			pixelsPerMM = float64(plan.PixelHeight) / plan.RenderedHeightMM
		}
		radius := int(blurMM*pixelsPerMM + 0.5)
		if radius < 1 {
			radius = 1
		}
		blurredPath = filepath.Join(dir, "blurred.png")
		if err := writeCoverPNG(blurredPath, boxBlurCoverImage(treated, radius)); err != nil {
			return preparedCoverVariants{}, err
		}
	}
	backdropPath := ""
	if plan.Config.BackdropMode == "image" {
		originalPath := plan.OriginalSourcePath
		if originalPath == "" {
			originalPath = plan.SourcePath
		}
		original, originalFormat, err := decodeCoverImage(originalPath)
		if err != nil {
			return preparedCoverVariants{}, err
		}
		if originalFormat != "png" && originalFormat != "jpeg" {
			return preparedCoverVariants{}, fmt.Errorf("folio.cover-image.path has unsupported image format %q", originalFormat)
		}
		backdrop := treatCoverImage(original, plan.Config.ImageTreatment, plan.Config.ImageOpacity)
		backdrop = centreCropCoverImage(backdrop, plan.PageWidthMM/plan.PageHeightMM)
		backdrop = resizeCoverImageWithin(backdrop, 512)
		blurMM, err := millimetres(plan.Config.BackdropImageBlur)
		if err != nil {
			return preparedCoverVariants{}, fmt.Errorf("folio.cover-image.backdrop.image.blur: %w", err)
		}
		pixelsPerMM := math.Max(float64(backdrop.Bounds().Dx())/plan.PageWidthMM, float64(backdrop.Bounds().Dy())/plan.PageHeightMM)
		radius := int(blurMM*pixelsPerMM + 0.5)
		if radius < 1 {
			radius = 1
		}
		backdropPath = filepath.Join(dir, "backdrop.jpg")
		if err := writeCoverJPEG(backdropPath, boxBlurCoverImage(backdrop, radius)); err != nil {
			return preparedCoverVariants{}, err
		}
	}
	return preparedCoverVariants{ImagePath: treatedPath, BlurredPath: blurredPath, BackdropPath: backdropPath}, nil
}

func decodeCoverImage(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("opening folio.cover-image.path: %w", err)
	}
	decoded, format, decodeErr := image.Decode(file)
	closeErr := file.Close()
	if decodeErr != nil {
		return nil, "", fmt.Errorf("decoding folio.cover-image.path: %w", decodeErr)
	}
	if closeErr != nil {
		return nil, "", fmt.Errorf("closing folio.cover-image.path: %w", closeErr)
	}
	return decoded, format, nil
}

func centreCropCoverImage(source *image.NRGBA, targetRatio float64) *image.NRGBA {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	cropWidth, cropHeight := width, height
	if float64(width)/float64(height) > targetRatio {
		cropWidth = maxInt(1, int(float64(height)*targetRatio+0.5))
	} else {
		cropHeight = maxInt(1, int(float64(width)/targetRatio+0.5))
	}
	left, top := (width-cropWidth)/2, (height-cropHeight)/2
	result := image.NewNRGBA(image.Rect(0, 0, cropWidth, cropHeight))
	for y := 0; y < cropHeight; y++ {
		copy(result.Pix[y*result.Stride:y*result.Stride+cropWidth*4], source.Pix[(top+y)*source.Stride+left*4:(top+y)*source.Stride+(left+cropWidth)*4])
	}
	return result
}

func resizeCoverImageWithin(source *image.NRGBA, longestLimit int) *image.NRGBA {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	longest := maxInt(width, height)
	if longest <= longestLimit {
		return source
	}
	scale := float64(longestLimit) / float64(longest)
	targetWidth := maxInt(1, int(float64(width)*scale+0.5))
	targetHeight := maxInt(1, int(float64(height)*scale+0.5))
	result := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		sourceY := (float64(y)+0.5)*float64(height)/float64(targetHeight) - 0.5
		for x := 0; x < targetWidth; x++ {
			sourceX := (float64(x)+0.5)*float64(width)/float64(targetWidth) - 0.5
			result.SetNRGBA(x, y, bilinearCoverPixel(source, sourceX, sourceY))
		}
	}
	return result
}

func bilinearCoverPixel(source *image.NRGBA, x, y float64) color.NRGBA {
	bounds := source.Bounds()
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	x1, y1 := x0+1, y0+1
	x0, x1 = minInt(maxInt(x0, 0), bounds.Dx()-1), minInt(maxInt(x1, 0), bounds.Dx()-1)
	y0, y1 = minInt(maxInt(y0, 0), bounds.Dy()-1), minInt(maxInt(y1, 0), bounds.Dy()-1)
	xWeight, yWeight := x-math.Floor(x), y-math.Floor(y)
	p00, p10 := source.NRGBAAt(x0, y0), source.NRGBAAt(x1, y0)
	p01, p11 := source.NRGBAAt(x0, y1), source.NRGBAAt(x1, y1)
	interpolate := func(a, b, c, d uint8) uint8 {
		top := float64(a)*(1-xWeight) + float64(b)*xWeight
		bottom := float64(c)*(1-xWeight) + float64(d)*xWeight
		return uint8(top*(1-yWeight) + bottom*yWeight + 0.5)
	}
	return color.NRGBA{
		R: interpolate(p00.R, p10.R, p01.R, p11.R),
		G: interpolate(p00.G, p10.G, p01.G, p11.G),
		B: interpolate(p00.B, p10.B, p01.B, p11.B),
		A: interpolate(p00.A, p10.A, p01.A, p11.A),
	}
}

func treatCoverImage(source image.Image, treatment string, opacity float64) *image.NRGBA {
	bounds := source.Bounds()
	result := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	backing := 255.0
	if treatment == "dark" {
		backing = 0
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			alpha := float64(pixel.A) / 255 * opacity
			composite := func(value uint8) uint8 {
				return uint8((float64(value)*alpha + backing*(1-alpha)) + 0.5)
			}
			result.SetNRGBA(x-bounds.Min.X, y-bounds.Min.Y, color.NRGBA{R: composite(pixel.R), G: composite(pixel.G), B: composite(pixel.B), A: 255})
		}
	}
	return result
}

func boxBlurCoverImage(source *image.NRGBA, radius int) *image.NRGBA {
	if radius < 1 {
		clone := image.NewNRGBA(source.Bounds())
		copy(clone.Pix, source.Pix)
		return clone
	}
	bounds := source.Bounds()
	row := make([]byte, bounds.Dx()*4)
	for y := 0; y < bounds.Dy(); y++ {
		var sums [4]int
		for sample := 0; sample <= minInt(bounds.Dx()-1, radius); sample++ {
			offset := y*source.Stride + sample*4
			for channel := range sums {
				sums[channel] += int(source.Pix[offset+channel])
			}
		}
		for x := 0; x < bounds.Dx(); x++ {
			left, right := maxInt(0, x-radius), minInt(bounds.Dx()-1, x+radius)
			offset := x * 4
			for channel := range sums {
				row[offset+channel] = uint8(sums[channel] / (right - left + 1))
			}
			if remove := x - radius; remove >= 0 {
				offset = y*source.Stride + remove*4
				for channel := range sums {
					sums[channel] -= int(source.Pix[offset+channel])
				}
			}
			if add := x + radius + 1; add < bounds.Dx() {
				offset = y*source.Stride + add*4
				for channel := range sums {
					sums[channel] += int(source.Pix[offset+channel])
				}
			}
		}
		copy(source.Pix[y*source.Stride:y*source.Stride+bounds.Dx()*4], row)
	}
	result := image.NewNRGBA(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		var sums [4]int
		for sample := 0; sample <= minInt(bounds.Dy()-1, radius); sample++ {
			offset := sample*source.Stride + x*4
			for channel := range sums {
				sums[channel] += int(source.Pix[offset+channel])
			}
		}
		for y := 0; y < bounds.Dy(); y++ {
			top, bottom := maxInt(0, y-radius), minInt(bounds.Dy()-1, y+radius)
			offset := y*result.Stride + x*4
			for channel := range sums {
				result.Pix[offset+channel] = uint8(sums[channel] / (bottom - top + 1))
			}
			if remove := y - radius; remove >= 0 {
				offset = remove*source.Stride + x*4
				for channel := range sums {
					sums[channel] -= int(source.Pix[offset+channel])
				}
			}
			if add := y + radius + 1; add < bounds.Dy() {
				offset = add*source.Stride + x*4
				for channel := range sums {
					sums[channel] += int(source.Pix[offset+channel])
				}
			}
		}
	}
	return result
}

func writeCoverPNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating prepared folio.cover-image: %w", err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		return fmt.Errorf("encoding prepared folio.cover-image: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing prepared folio.cover-image: %w", err)
	}
	return nil
}

func writeCoverJPEG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating prepared folio.cover-image backdrop: %w", err)
	}
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: coverBackdropJPEGQuality}); err != nil {
		_ = file.Close()
		return fmt.Errorf("encoding prepared folio.cover-image backdrop: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing prepared folio.cover-image backdrop: %w", err)
	}
	return nil
}

func coverDecodeConfig(path string) (uint32, uint32, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, "", err
	}
	decoded, format, err := image.DecodeConfig(file)
	closeErr := file.Close()
	if err != nil {
		return 0, 0, "", err
	}
	if closeErr != nil {
		return 0, 0, "", closeErr
	}
	if decoded.Width < 0 || decoded.Height < 0 {
		return 0, 0, "", fmt.Errorf("negative image dimensions")
	}
	return uint32(decoded.Width), uint32(decoded.Height), format, nil
}

func validateDodgeGroupFit(group string, measured, available float64) error {
	if measured > available {
		return fmt.Errorf("folio.cover-image dodge %s group height %.1fpt exceeds available height %.1fpt", group, measured, available)
	}
	return nil
}

func millimetres(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if !strings.HasSuffix(value, "mm") {
		return 0, fmt.Errorf("cover length %q is not in mm", value)
	}
	result, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "mm")), 64)
	if err != nil {
		return 0, fmt.Errorf("parsing cover length %q: %w", value, err)
	}
	return result, nil
}

func coverPageDimensionsMM(page string) (float64, float64) {
	switch strings.ToLower(strings.TrimSpace(page)) {
	case "a5":
		return 148, 210
	case "letter", "us-letter":
		return 215.9, 279.4
	case "legal", "us-legal":
		return 215.9, 355.6
	default:
		return 210, 297
	}
}

func maxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
