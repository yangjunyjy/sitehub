package services

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestPuzzleMatchesOriginalPixelsAndSlotOutline(t *testing.T) {
	background := image.NewRGBA(image.Rect(0, 0, captchaImgWidth, captchaImgHeight))
	for y := 0; y < captchaImgHeight; y++ {
		for x := 0; x < captchaImgWidth; x++ {
			background.SetRGBA(x, y, color.RGBA{R: uint8(x % 200), G: uint8(y), B: 120, A: 255})
		}
	}
	x, y := 137, 46
	piece := cropSliderPiece(background, x, y)
	if piece.Bounds().Dx() != puzzleSize+2*piecePadding || piece.Bounds().Dy() != puzzleSize+2*piecePadding {
		t.Fatal("piece dimensions do not include symmetric padding")
	}
	original := append([]byte(nil), background.Pix...)
	drawSlot(background, x, y)
	for dy := 0; dy < puzzleSize; dy++ {
		for dx := 0; dx < puzzleSize; dx++ {
			pixel := piece.RGBAAt(piecePadding+dx, piecePadding+dy)
			before := color.RGBA{R: uint8((x + dx) % 200), G: uint8(y + dy), B: 120, A: 255}
			if !insidePuzzle(dx, dy) {
				if pixel.A != 0 || background.RGBAAt(x+dx, y+dy) != before {
					t.Fatal("piece transparency and slot outline differ")
				}
			} else if !puzzleEdge(dx, dy) {
				if pixel != before {
					t.Fatalf("piece sampled wrong source coordinate at %d,%d: %v != %v", dx, dy, pixel, before)
				}
				if background.RGBAAt(x+dx, y+dy) == before {
					t.Fatal("slot interior was not darkened")
				}
			}
		}
	}
	if bytes.Equal(original, background.Pix) || piece.RGBAAt(0, 0).A != 0 {
		t.Fatal("missing slot or transparent padding")
	}
}

func TestCaptchaResponseDoesNotExposeAnswer(t *testing.T) {
	data, err := json.Marshal(SliderCaptchaResult{ImgWidth: captchaImgWidth, ImgHeight: captchaImgHeight, PieceSize: puzzleSize, PiecePadding: piecePadding})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if _, exists := payload["x"]; exists {
		t.Fatal("challenge reveals target position")
	}
	for _, field := range []string{"imgWidth", "imgHeight", "pieceSize", "piecePadding"} {
		if payload[field].(float64) <= 0 {
			t.Fatalf("missing image geometry: %s", field)
		}
	}
}

func TestCaptchaUnavailableCacheReturnsError(t *testing.T) {
	service := NewSliderCaptchaService(nil)
	if _, err := service.Generate(context.Background()); err == nil {
		t.Fatal("expected cache error")
	}
	if _, err := service.Verify(context.Background(), "key", 120, 600, 10); err == nil {
		t.Fatal("expected cache error")
	}
}

func TestCaptchaConsumesChallengeOnce(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := NewSliderCaptchaService(client)
	ctx := context.Background()
	generate := func() (*SliderCaptchaResult, sliderCaptchaData) {
		t.Helper()
		result, err := service.Generate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := server.Get(sliderCaptchaPrefix + result.CaptchaKey)
		if err != nil {
			t.Fatal(err)
		}
		var answer sliderCaptchaData
		if err := json.Unmarshal([]byte(stored), &answer); err != nil {
			t.Fatal(err)
		}
		return result, answer
	}

	result, answer := generate()
	var succeeded atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ok, err := service.Verify(ctx, result.CaptchaKey, answer.X, 700, 12)
			if err != nil {
				t.Errorf("verification failed: %v", err)
			}
			if ok {
				succeeded.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if succeeded.Load() != 1 {
		t.Fatalf("concurrent successful uses = %d, want 1", succeeded.Load())
	}

	for _, test := range []struct {
		name             string
		offset, duration int
		expire           bool
	}{
		{name: "wrong position", offset: 20, duration: 700},
		{name: "too fast", duration: 10},
		{name: "expired", duration: 700, expire: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, answer := generate()
			if test.expire {
				server.FastForward(sliderCaptchaTTL + time.Second)
			}
			if ok, err := service.Verify(ctx, result.CaptchaKey, answer.X+test.offset, test.duration, 10); err != nil || ok {
				t.Fatalf("invalid attempt accepted or failed unexpectedly: %v, %v", ok, err)
			}
			if ok, err := service.Verify(ctx, result.CaptchaKey, answer.X, 700, 10); err != nil || ok {
				t.Fatalf("consumed challenge reused: %v, %v", ok, err)
			}
		})
	}
}
