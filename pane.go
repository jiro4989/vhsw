package main

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

type Pane struct {
	Name    string
	X       int
	Y       int
	Width   int
	Height  int
	Command string
}

type Panes []Pane

func NewPanes(c, tw, th int, cmds []string) (ret Panes) {
	cmdLen := len(cmds)
	for i, cmd := range cmds {
		pw := paneWidth(c, tw)          // 1ペインあたりの幅
		ph := paneHeight(c, th, cmdLen) // 1ペインあたりの幅

		x := paneX(c, pw, i)
		y := paneY(c, ph, i)
		// ペインの下のペインが空いていたら縦幅を拡張
		h := ph
		if cmdLen-1 < i+c && (cmdLen-1)/c != i/c {
			h += ph
		}
		p := Pane{
			Name:    cmd,
			X:       x,
			Y:       y,
			Width:   pw,
			Height:  h,
			Command: cmd,
		}
		ret = append(ret, p)
	}
	return
}

func paneWidth(c, w int) int {
	return w / c
}

func paneHeight(c, h, cnt int) int {
	mod := cnt % c
	div := cnt / c
	if 0 < mod {
		div++
	}
	return h / div
}

func paneX(c, w, i int) int {
	return w * (i % c)
}

func paneY(c, h, i int) int {
	return h * (i / c)
}

// DrawHeader はヘッダ情報を描画する。
// ヘッダはコマンド名を表示し、背景色を変更する。
func (p *Pane) DrawHeader() {
	// ヘッダの背景色を変更
	w, _ := termbox.Size()
	bgline := strings.Repeat(" ", w)
	p.DrawLineText(0, p.Y, bgline, termbox.ColorBlack, termbox.ColorWhite, true)

	// 上書きでテキストをセット
	now := time.Now().Format("2006/01/02 15:04:05")
	line := p.Name + " " + now
	p.DrawLineText(0, p.Y, line, termbox.ColorBlack, termbox.ColorWhite, true)
}

// DrawText はテキストをペインにセットする。
// セット対象のテキストがペインの表示領域を超過しそうな場合は
// 超過しないように切り落とす。
// termbox.Flushしないので、別途Flushが必要
func (p *Pane) DrawText(x, y int, b []byte, fc, bc termbox.Attribute, chopLongLines bool) {
	s := string(b)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		i += y
		if p.Height < i {
			break
		}
		fy := p.Y + i
		p.DrawLineText(x, fy, line, fc, bc, chopLongLines)
		// 折り返しが発生していたときに、折り返された行数分ずらす
		if !chopLongLines {
			w := runewidth.StringWidth(line)
			y += w / p.Width
		}
	}
}

// DrawLineText(y は１行のテキストをターミナルに書く。
// １行がペインの幅に収まりきらないときに、切り詰めるか、折り返すかを
// chopLongLinesで管理する。
// termbox.Flushはしない。
func (p *Pane) DrawLineText(x, y int, line string, fc, bc termbox.Attribute, chopLongLines bool) {
	var colPos int
	for _, c := range []rune(line) {
		colPos += x
		if p.Width < colPos {
			if chopLongLines {
				// はみ出してしまっていたときは
				// テキストが切り落とされていることを明示する
				termbox.SetCell(p.X+colPos-1, y, '.', fc, bc)
				termbox.SetCell(p.X+colPos-2, y, '.', fc, bc)
				break
			}
			colPos = 0
			y++
		}
		fx := p.X + colPos
		termbox.SetCell(fx, y, c, fc, bc)
		l := runewidth.StringWidth(string(c))
		// マルチバイト文字を処理したときは1文字ずらす
		if 1 < l {
			termbox.SetCell(fx+1, y, ' ', fc, bc)
			colPos++
		}
		colPos++
	}
}
