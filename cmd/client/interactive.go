package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/nsf/termbox-go"
	"github.com/ozeidan/gosearch/pkg/client"
)

type point struct {
	x int
	y int
}

type booleanOpt struct {
	name   string
	active bool
	option client.Option
}

type uiState struct {
	query     string
	results   []string
	scrollPos point
}

const (
	checkBoxUnchecked = '⚪'
	checkBoxChecked   = '⚫'
	queryString       = "query: "
)

const (
	headerHeight   = 2
	settingsMargin = 3
)

var cursorStart = point{len(queryString), 0}

var fuzzyOption = booleanOpt{name: "Fuzzy", option: client.Fuzzy}
var prefixOption = booleanOpt{name: "Prefix", option: client.PrefixSearch}
var substringOption = booleanOpt{name: "Substring", option: client.SubStringSearch}

var searchOptions = []*booleanOpt{
	&fuzzyOption,
	&prefixOption,
	&substringOption,
}

var current uiState = uiState{}

func startInteractive() error {
	err := termbox.Init()

	if err != nil {
		return err
	}

	defer termbox.Close()
	renderUI()
	termbox.Sync()
	eventLoop()
	return nil
}

func eventLoop() {
	for {
		event := termbox.PollEvent()

		if event.Type != termbox.EventKey {
			if event.Type == termbox.EventResize {
				renderUI()
				termbox.Sync()
			}
			continue
		}

		switch event.Key {
		case termbox.KeyEsc:
			fallthrough
		case termbox.KeyCtrlC:
			return
		case termbox.KeyCtrlF:
			setSearchMode(&fuzzyOption)
		case termbox.KeyCtrlP:
			setSearchMode(&prefixOption)
		case termbox.KeyCtrlS:
			setSearchMode(&substringOption)
		case termbox.KeyEnter:
			runQuery()
		case termbox.KeyBackspace:
			fallthrough
		case termbox.KeyBackspace2:
			deleteChar(1)
			renderQuery()
		case termbox.KeyCtrlW:
			deleteLastWord()
			renderQuery()
		case termbox.KeySpace:
			event.Ch = ' '
			fallthrough
		default:
			if event.Ch == 0 {
				continue
			}

			setChar(event.Ch)
			if len(current.query) > 3 {
				runQuery()
			}
			renderQuery()
		}

		termbox.Sync()
	}
}

func setSearchMode(opt *booleanOpt) {
	for _, searchOption := range searchOptions {
		if opt == searchOption {
			searchOption.active = true
		} else {
			searchOption.active = false
		}
	}
	renderSettings()
}

func runQuery() {
	if current.query == "" {
		return
	}

	opts := []client.Option{client.ReverseSort}
	for _, opt := range searchOptions {
		if opt.active {
			opts = append(opts, opt.option)
		}
	}
	current.results = current.results[:0]

	go func() {
		responseChan, err := client.SearchRequest(current.query, opts...)
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			return
		}

		for response := range responseChan {
			current.results = append(current.results, response)
		}
		renderResults()
		termbox.Sync()
	}()
}

func deleteLastWord() {
	q := []rune(current.query)
	if len(q) == 0 {
		return
	}

	foundNonWhiteSpace := false
	i := len(q) - 1
	for ; i > 0; i-- {
		if !unicode.IsSpace(q[i]) {
			foundNonWhiteSpace = true
		} else if foundNonWhiteSpace {
			break
		}
	}

	q = q[:i]
	current.query = string(q)
}

func setChar(c rune) {
	current.query += string(c)
}

func deleteChar(amount int) {
	current.query = current.query[:len(current.query)-1]
}

func writeString(s string, x, y int) {
	for _, r := range s {
		termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
		x++
	}
}

func writeStringRev(s string, x, y int, fgColor, bgColor termbox.Attribute) {
	runes := []rune(s)
	for xs := len(runes) - 1; x >= 0 && xs >= 0; {
		termbox.SetCell(x, y, runes[xs], fgColor, bgColor)
		x--
		xs--
	}
}

func renderUI() {
	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
	fmt.Printf("current = %+v\n", current)

	renderQuery()
	renderSettings()
	renderResults()
}

func renderQuery() {
	writeString(queryString, 0, 0)
	writeString(current.query, cursorStart.x, cursorStart.y)
	termWidth, _ := termbox.Size()
	clearArea(cursorStart.x+len(current.query), 0, termWidth, 0)
	termbox.SetCursor(cursorStart.x+len(current.query), 0)
}

func renderSettings() {
	termWidth, _ := termbox.Size()
	setArea(0, 1, termWidth, 1, termbox.ColorWhite)

	settingsString := strings.Repeat(" ", settingsMargin)
	for _, opt := range searchOptions {
		r := modeToRune(opt.active)
		optionString := fmt.Sprintf("%s: %c ", opt.name, r)
		settingsString = optionString + settingsString
	}

	writeStringRev(settingsString, termWidth-1, 1, termbox.ColorBlack, termbox.ColorWhite)
}

func modeToRune(mode bool) rune {
	if mode {
		return checkBoxChecked
	} else {
		return checkBoxUnchecked
	}
}

func renderResults() {
	termWidth, termHeight := termbox.Size()
	results := current.results[current.scrollPos.y:]
	y := headerHeight

	for _, result := range results {
		if y >= termHeight {
			break
		}
		x := 0
		for _, r := range result {
			if x >= termWidth {
				break
			}

			termbox.SetCell(x, y, r, termbox.ColorWhite, termbox.ColorBlack)
			x++
		}
		clearArea(x, y, termWidth, y)
		y++
	}

	clearArea(0, y, termWidth, termHeight)
}

func clearArea(xStart, yStart, xEnd, yEnd int) {
	setArea(xStart, yStart, xEnd, yEnd, termbox.ColorBlack)
}

func setArea(xStart, yStart, xEnd, yEnd int, color termbox.Attribute) {
	for x := xStart; x <= xEnd; x++ {
		for y := yStart; y <= yEnd; y++ {
			termbox.SetCell(x, y, ' ', termbox.ColorWhite, color)
		}
	}
}
