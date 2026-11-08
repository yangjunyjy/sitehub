package services

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	sliderCaptchaPrefix = "slider_captcha:"
	sliderCaptchaTTL    = 5 * time.Minute
)

// SliderCaptchaService 滑块验证码服务接口
type SliderCaptchaService interface {
	// Generate 生成拼图验证码（返回背景图、拼图块图、captchaKey、y位置）
	Generate(ctx context.Context) (*SliderCaptchaResult, error)
	// Verify 校验滑块位置（含人类行为检测）
	Verify(ctx context.Context, captchaKey string, sliderX int, duration int, moveCount int) (bool, error)
}

type sliderCaptchaService struct {
	redisClient *redis.Client
}

// NewSliderCaptchaService 创建滑块验证码服务
func NewSliderCaptchaService(redisClient *redis.Client) SliderCaptchaService {
	return &sliderCaptchaService{redisClient: redisClient}
}

// SliderCaptchaResult 生成结果
type SliderCaptchaResult struct {
	CaptchaKey string `json:"captchaKey"`
	// 背景图 base64（带缺口）
	BackgroundImage string `json:"backgroundImage"`
	// 拼图块图 base64（带阴影）
	SliderImage string `json:"sliderImage"`
	// 拼图块Y坐标（提供给前端定位）
	Y int `json:"y"`
	// 图片原始宽度（供前端计算缩放比例）
	ImgWidth     int `json:"imgWidth"`
	ImgHeight    int `json:"imgHeight"`
	PieceSize    int `json:"pieceSize"`
	PiecePadding int `json:"piecePadding"`
}

// sliderCaptchaData 存储在Redis中的验证码数据
type sliderCaptchaData struct {
	X      int `json:"x"`      // 缺口X坐标
	Y      int `json:"y"`      // 缺口Y坐标
	Width  int `json:"width"`  // 拼图块宽度
	Height int `json:"height"` // 拼图块高度
}

// 图片尺寸常量
const (
	captchaImgWidth  = 300
	captchaImgHeight = 150
	puzzleSize       = 44 // 拼图轮廓的边界尺寸
	piecePadding     = 4
)

// Generate 生成滑块拼图验证码
func (s *sliderCaptchaService) Generate(ctx context.Context) (*SliderCaptchaResult, error) {
	if s.redisClient == nil {
		return nil, errors.New("验证码缓存不可用")
	}
	// 1. 生成随机背景图
	bgImg := generateBackgroundImage(captchaImgWidth, captchaImgHeight)

	// 2. 随机生成缺口位置
	minX := 100
	maxX := captchaImgWidth - puzzleSize - 10
	x := rand.Intn(maxX-minX) + minX
	y := rand.Intn(captchaImgHeight-puzzleSize-10) + 5

	// 先从原始图裁剪，再用同一轮廓挖空，确保拼图颜色与缺口准确对应。
	sliderImg := cropSliderPiece(bgImg, x, y)
	drawSlot(bgImg, x, y)

	// 5. 生成captchaKey
	captchaKey, err := generateCaptchaKey()
	if err != nil {
		return nil, fmt.Errorf("生成验证码标识失败: %w", err)
	}

	// 6. 存储到Redis
	data := &sliderCaptchaData{
		X:      x,
		Y:      y,
		Width:  puzzleSize,
		Height: puzzleSize,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化验证码数据失败: %w", err)
	}

	if err := s.redisClient.Set(ctx, sliderCaptchaPrefix+captchaKey, string(dataJSON), sliderCaptchaTTL).Err(); err != nil {
		return nil, fmt.Errorf("存储验证码到Redis失败: %w", err)
	}

	// 7. 编码为base64
	bgBase64, err := encodeImageToBase64(bgImg, "png")
	if err != nil {
		return nil, fmt.Errorf("编码背景图失败: %w", err)
	}

	sliderBase64, err := encodeImageToBase64(sliderImg, "png")
	if err != nil {
		return nil, fmt.Errorf("编码拼图块失败: %w", err)
	}

	return &SliderCaptchaResult{
		CaptchaKey:      captchaKey,
		BackgroundImage: "data:image/png;base64," + bgBase64,
		SliderImage:     "data:image/png;base64," + sliderBase64,
		Y:               y,
		ImgWidth:        captchaImgWidth,
		ImgHeight:       captchaImgHeight,
		PieceSize:       puzzleSize,
		PiecePadding:    piecePadding,
	}, nil
}

// Verify 校验滑块位置（含人类行为检测）
func (s *sliderCaptchaService) Verify(ctx context.Context, captchaKey string, sliderX int, duration int, moveCount int) (bool, error) {
	if s.redisClient == nil {
		return false, errors.New("验证码缓存不可用")
	}
	// 原子取出并删除，避免并发提交重复使用同一张拼图；兼容旧版 Redis。
	val, err := s.redisClient.Eval(ctx, `
local value = redis.call('GET', KEYS[1])
if value then redis.call('DEL', KEYS[1]) end
return value
`, []string{sliderCaptchaPrefix + captchaKey}).Text()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取验证码失败: %w", err)
	}
	if !isHumanBehavior(duration, moveCount) {
		return false, nil
	}
	var data sliderCaptchaData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return false, fmt.Errorf("解析验证码数据失败: %w", err)
	}
	return sliderX >= data.X-5 && sliderX <= data.X+5, nil
}

// 人类行为检测阈值
const (
	minDragDuration = 300   // 最短拖动时长(ms) —— 低于此值视为机器人
	maxDragDuration = 10000 // 最长拖动时长(ms) —— 超过视为超时/非人类
	minMoveCount    = 3     // 最少移动轨迹点数 —— 低于此值视为机器人直线拖拽
)

// isHumanBehavior 检测拖动行为是否符合人类特征
func isHumanBehavior(duration, moveCount int) bool {
	// 时长校验：太快说明是脚本，太慢说明不是正常操作
	if duration < minDragDuration || duration > maxDragDuration {
		return false
	}
	// 轨迹点数校验：人类拖动会产生大量中间移动点，机器人通常只有极少量的点
	// 300ms内浏览器 mousemove 事件通常触发 20-60 次，3次是绝对最低下限
	if moveCount < minMoveCount {
		return false
	}
	return true
}

// generateBackgroundImage 绘制有层次的山景，给拼图提供清晰可辨的纹理。
func generateBackgroundImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	palettes := [][5]color.RGBA{
		{rgba(198, 227, 235, 255), rgba(239, 244, 227, 255), rgba(134, 177, 185, 255), rgba(81, 135, 146, 255), rgba(42, 90, 110, 255)},
		{rgba(218, 218, 242, 255), rgba(253, 232, 211, 255), rgba(173, 166, 204, 255), rgba(118, 129, 171, 255), rgba(65, 88, 130, 255)},
		{rgba(216, 234, 217, 255), rgba(247, 244, 209, 255), rgba(155, 185, 157, 255), rgba(98, 149, 130, 255), rgba(46, 105, 103, 255)},
	}
	palette := palettes[rand.Intn(len(palettes))]
	phase := rand.Float64() * math.Pi * 2
	sunX, sunY := width/2+rand.Intn(width/3), height/5+rand.Intn(height/6)
	mix := func(a, b uint8, t float64) uint8 { return uint8(float64(a)*(1-t) + float64(b)*t) }
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			t := float64(y) / float64(height)
			c := rgba(mix(palette[0].R, palette[1].R, t), mix(palette[0].G, palette[1].G, t), mix(palette[0].B, palette[1].B, t), 255)
			if (x-sunX)*(x-sunX)+(y-sunY)*(y-sunY) < 16*16 {
				c = rgba(255, 247, 218, 255)
			}
			for layer := 0; layer < 3; layer++ {
				ridge := float64(height)*(0.47+float64(layer)*0.17) +
					math.Sin(float64(x)/49+phase+float64(layer)*1.8)*17 +
					math.Sin(float64(x)/19+float64(layer))*5
				if float64(y) >= ridge {
					c = palette[layer+2]
					shade := 0.96 + 0.04*math.Sin(float64(x+y)/22+phase)
					c.R, c.G, c.B = uint8(float64(c.R)*shade), uint8(float64(c.G)*shade), uint8(float64(c.B)*shade)
				}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// 拼图与缺口共享轮廓；凸起和凹口均位于固定边界内，不裁断图形。
func insidePuzzle(x, y int) bool {
	if x < 0 || y < 0 || x >= puzzleSize || y >= puzzleSize {
		return false
	}
	circle := func(cx, cy, radius int) bool { return (x-cx)*(x-cx)+(y-cy)*(y-cy) <= radius*radius }
	body := x >= 1 && x < puzzleSize-8 && y >= 8 && y < puzzleSize-1
	shape := body || circle(21, 8, 7) || circle(puzzleSize-8, 27, 7)
	return shape && !circle(1, 28, 6)
}

func puzzleEdge(x, y int) bool {
	return !insidePuzzle(x-1, y) || !insidePuzzle(x+1, y) || !insidePuzzle(x, y-1) || !insidePuzzle(x, y+1)
}

func drawSlot(bg *image.RGBA, x, y int) {
	for dy := 0; dy < puzzleSize; dy++ {
		for dx := 0; dx < puzzleSize; dx++ {
			if !insidePuzzle(dx, dy) {
				continue
			}
			c := bg.RGBAAt(x+dx, y+dy)
			if puzzleEdge(dx, dy) {
				c = rgba(uint8((int(c.R)+255)/2), uint8((int(c.G)+255)/2), uint8((int(c.B)+255)/2), 255)
			} else {
				c = rgba(uint8(float64(c.R)*0.42), uint8(float64(c.G)*0.42), uint8(float64(c.B)*0.42), 255)
			}
			bg.SetRGBA(x+dx, y+dy, c)
		}
	}
}

func cropSliderPiece(bg *image.RGBA, x, y int) *image.RGBA {
	piece := image.NewRGBA(image.Rect(0, 0, puzzleSize+piecePadding*2, puzzleSize+piecePadding*2))
	for dy := 0; dy < puzzleSize; dy++ {
		for dx := 0; dx < puzzleSize; dx++ {
			if !insidePuzzle(dx, dy) {
				continue
			}
			c := bg.RGBAAt(x+dx, y+dy)
			if puzzleEdge(dx, dy) {
				c = rgba(255, 255, 255, 255)
			}
			piece.SetRGBA(piecePadding+dx, piecePadding+dy, c)
		}
	}
	return piece
}

func encodeImageToBase64(img image.Image, format string) (string, error) {
	var buf bytes.Buffer
	var err error

	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(&buf, img)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, nil)
	default:
		err = png.Encode(&buf, img)
	}

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func generateCaptchaKey() (string, error) {
	var value [24]byte
	if _, err := cryptorand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func rgba(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}
