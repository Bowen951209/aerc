package app

import (
	"errors"
	"fmt"
	"strings"

	"git.sr.ht/~rjarry/aerc/config"
	"git.sr.ht/~rjarry/aerc/lib/authres"
	"git.sr.ht/~rjarry/aerc/lib/ui"
	"github.com/mattn/go-runewidth"
	"go.rockorager.dev/vaxis"
)

var ErrNoHeader = errors.New("(no header)")

type Chunk interface {
	isChunk()
}

func chunkToString(c Chunk) string {
	var s string

	switch c := c.(type) {
	case ChunkResult:
		s = c.Result.Symbol()
	case ChunkText:
		s = c.Text
	}

	return s
}

func chunksToString(chunks []Chunk) string {
	var sb strings.Builder
	for _, chunk := range chunks {
		sb.WriteString(chunkToString(chunk))
	}

	return sb.String()
}

type ChunkResult struct {
	Result authres.Result
}

func (ChunkResult) isChunk() {}

type ChunkText struct {
	Text string
}

func (ChunkText) isChunk() {}

type AuthInfo struct {
	authdetails *authres.Details
	showInfo    bool
	uiConfig    *config.UIConfig
}

func NewAuthInfo(auth *authres.Details, showInfo bool, uiConfig *config.UIConfig) *AuthInfo {
	return &AuthInfo{authdetails: auth, showInfo: showInfo, uiConfig: uiConfig}
}

func (a *AuthInfo) Draw(ctx *ui.Context) {
	defaultStyle := a.uiConfig.GetStyle(config.STYLE_DEFAULT)
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', defaultStyle)

	chunks, err := FormatAuthInfoToChunks(a.authdetails, a.showInfo, ctx.Width()-1)

	var style vaxis.Style
	if err != nil {
		if errors.Is(err, ErrNoHeader) {
			style = defaultStyle
		} else {
			style = a.uiConfig.GetStyle(config.STYLE_ERROR)
		}
		ctx.Printf(0, 0, style, "%s", err)
		return
	}

	x := 1
	for _, chunk := range chunks {
		var style vaxis.Style

		switch c := chunk.(type) {
		case ChunkText:
			style = defaultStyle
		case ChunkResult:
			switch c.Result {
			case authres.ResultNone:
				style = defaultStyle
			case authres.ResultNeutral:
				style = a.uiConfig.GetStyle(config.STYLE_WARNING)
			case authres.ResultPolicy:
				style = a.uiConfig.GetStyle(config.STYLE_WARNING)
			case authres.ResultPass:
				style = a.uiConfig.GetStyle(config.STYLE_SUCCESS)
			case authres.ResultFail:
				style = a.uiConfig.GetStyle(config.STYLE_ERROR)
			default:
				style = a.uiConfig.GetStyle(config.STYLE_ERROR)
			}
		}

		text := chunkToString(chunk)
		ctx.Printf(x, 0, style, "%s", text)
		x += runewidth.StringWidth(text)
	}
}

func (a *AuthInfo) Invalidate() {
	ui.Invalidate()
}

func FormatAuthInfoToChunks(auth *authres.Details, showInfo bool, maxLen int) ([]Chunk, error) {
	if auth == nil {
		return nil, ErrNoHeader
	}
	if auth.Err != nil {
		return nil, auth.Err
	}

	// Build result symbols (e.g. ✓ ✗ ✓)
	var chunks []Chunk
	for i, res := range auth.Results {
		if i > 0 {
			chunks = append(chunks, ChunkText{" "})
		}
		chunks = append(chunks, ChunkResult{res})
	}

	if showInfo {
		infoText := ""
		for i, info := range auth.Infos {
			if i > 0 {
				infoText += ","
			}
			infoText += info + auth.Reasons[i]
		}

		if infoText != "" {
			availableWidth := maxLen - runewidth.StringWidth(chunksToString(chunks)) - 3
			if availableWidth > 0 {
				truncatedInfo := runewidth.Truncate(infoText, availableWidth, "…")
				infoText = fmt.Sprintf(" (%s)", truncatedInfo)
				chunks = append(chunks, ChunkText{infoText})
			}
		}
	}

	return chunks, nil
}
