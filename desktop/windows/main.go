//go:build windows

// Path Home: Structural Reentry Assistance — native Win32 application. Windows 10/11 x64; no browser required.
// Current product: Path Home. Original in-app photo assets and local data are preserved.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"log"
	"runtime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/reentry-public.json assets/photos/*.bmp assets/icons/*.bmp assets/PathHome.ico assets/mission.txt
var bundled embed.FS

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	shell32           = syscall.NewLazyDLL("shell32.dll")
	pRegisterClass    = user32.NewProc("RegisterClassExW")
	pCreateWindow     = user32.NewProc("CreateWindowExW")
	pDefWindowProc    = user32.NewProc("DefWindowProcW")
	pShowWindow       = user32.NewProc("ShowWindow")
	pUpdateWindow     = user32.NewProc("UpdateWindow")
	pGetMessage       = user32.NewProc("GetMessageW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
	pDispatchMessage  = user32.NewProc("DispatchMessageW")
	pPostQuit         = user32.NewProc("PostQuitMessage")
	pSendMessage      = user32.NewProc("SendMessageW")
	pSetText          = user32.NewProc("SetWindowTextW")
	pGetText          = user32.NewProc("GetWindowTextW")
	pGetTextLength    = user32.NewProc("GetWindowTextLengthW")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pMoveWindow       = user32.NewProc("MoveWindow")
	pMessageBox       = user32.NewProc("MessageBoxW")
	pLoadImage        = user32.NewProc("LoadImageW")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetBkColor       = gdi32.NewProc("SetBkColor")
	pInvalidateRect   = user32.NewProc("InvalidateRect")
	pSetTextColor     = gdi32.NewProc("SetTextColor")
	pCreateFont       = gdi32.NewProc("CreateFontW")
	pCreateBrush      = gdi32.NewProc("CreateSolidBrush")
	pCreatePen        = gdi32.NewProc("CreatePen")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pDeleteObject     = gdi32.NewProc("DeleteObject")
	pRoundRect        = gdi32.NewProc("RoundRect")
	pDrawText         = user32.NewProc("DrawTextW")
	pDrawFocusRect    = user32.NewProc("DrawFocusRect")
	pFillRect         = user32.NewProc("FillRect")
	pCopyMemory       = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlMoveMemory")
	pGetModule        = kernel32.NewProc("GetModuleHandleW")
	pShellExec        = shell32.NewProc("ShellExecuteW")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER           = 0x00800000
	WS_TABSTOP          = 0x00010000
	WS_VSCROLL          = 0x00200000
	ES_MULTILINE        = 0x0004
	ES_AUTOVSCROLL      = 0x0040
	ES_READONLY         = 0x0800
	BS_AUTOCHECKBOX     = 0x0003
	SS_BITMAP           = 0x000E
	WM_DESTROY          = 0x0002
	WM_CLOSE            = 0x0010
	WM_COMMAND          = 0x0111
	WM_CTLCOLORSTATIC   = 0x0138
	WM_CTLCOLOREDIT     = 0x0133
	WM_SETFONT          = 0x0030
	WM_GETTEXT          = 0x000D
	BM_GETCHECK         = 0x00F0
	BM_SETCHECK         = 0x00F1
	STM_SETIMAGE        = 0x0172
	IMAGE_BITMAP        = 0
	IMAGE_ICON          = 1
	LR_LOADFROMFILE     = 0x10
)

type point struct{ X, Y int32 }
type buttonRect struct{ Left, Top, Right, Bottom int32 }
type drawItem struct {
	CtlType, CtlID, ItemID, ItemAction, ItemState uint32
	HwndItem, DC uintptr
	Rect buttonRect
	ItemData uintptr
}
type message struct {
	Hwnd           uintptr
	Message        uint32
	_              uint32
	Wparam, Lparam uintptr
	Time           uint32
	Pt             point
	LPrivate       uint32
}
type wndClass struct {
	CbSize                          uint32
	Style                           uint32
	WndProc                         uintptr
	ClsExtra, WndExtra              int32
	Instance, Icon, Cursor, BgBrush uintptr
	MenuName, ClassName             *uint16
	SmallIcon                       uintptr
}
type state struct {
	Theme string `json:"theme,omitempty"`
	// Nil map denotes a legacy profile: zero-valued amounts are conservatively unknown.
	BudgetKnown map[string]bool `json:"budgetKnown,omitempty"`
	StayEnabled     bool            `json:"stayEnabled"`
	State           string          `json:"state"`
	County          string          `json:"county"`
	Zip             string          `json:"zip"`
	Release         string          `json:"release"`
	Start           string          `json:"start"`
	End             string          `json:"end"`
	Supervision     string          `json:"supervision"`
	Support         string          `json:"support"`
	AppointmentDate string          `json:"appointmentDate"`
	AppointmentText string          `json:"appointmentText"`
	Cash            float64         `json:"cash"`
	BirthCost       float64         `json:"birthCost"`
	IDCost          float64         `json:"idCost"`
	OtherCost       float64         `json:"otherCost"`
	Asked           string          `json:"asked"`
	UnstuckGoal     string          `json:"unstuckGoal,omitempty"`
	UnstuckBarrier  string          `json:"unstuckBarrier,omitempty"`
	HelpNeed        string          `json:"helpNeed,omitempty"`
	HelpPage        int             `json:"helpPage,omitempty"`
	Checklist       map[string]bool `json:"checklist"`
}
type legalRule struct {
	Jurisdiction   string `json:"jurisdiction_code"`
	Title          string `json:"rule_title"`
	Plain          string `json:"plain_language"`
	Qualifications string `json:"qualifications"`
	Source         string `json:"source_url"`
	Topic          string `json:"topic_code"`
	Supervision    string `json:"supervision_type"`
	Status         string `json:"legal_review_status"`
}
type program struct {
	Name        string `json:"program_name"`
	State       string `json:"state_code"`
	County      string `json:"county"`
	Categories  string `json:"service_categories"`
	Phone       string `json:"phone"`
	Website     string `json:"website"`
	Eligibility string `json:"eligibility_description"`
	Funding     string `json:"funding_status"`
}
type database struct {
	Rules          []legalRule     `json:"legal_rules"`
	Programs       []program       `json:"assistance_programs"`
	LocalProviders []localProvider `json:"local_providers"`
	Jurisdictions  []struct {
		Code string `json:"jurisdiction_code"`
		Name string `json:"jurisdiction_name"`
	} `json:"jurisdictions"`
}

var (
	hwnd                       uintptr
	font, largeFont, titleFont, buttonFont uintptr
	brush                      uintptr
	activePalette = palettes[0]
	// Retain the class background brush for the lifetime of the registered window class.
	classBrush uintptr
	controls                   []uintptr
	bitmaps                    []uintptr
	buttonLabels               map[uint32]string
	fields                     map[string]uintptr
	clickActions               map[int]func()
	input                      = state{Checklist: map[string]bool{}}
	data                       database
	page                       = "home"
	panelW                     = 1090
	pageW                      = 810
	assetDir                   string
	lastLink                   string
)

type visualPalette struct {
	ID, Name string
	Background, Ink, Primary, PrimaryInk, Secondary, SecondaryInk, Edge uintptr
}
// RGB values are converted to Win32 COLORREF (BGR). Accent colors never replace the Path Home logo.
func rgb(v uintptr) uintptr { return (v&0xff)<<16 | v&0xff00 | (v>>16)&0xff }
var palettes = []visualPalette{
	{"classic", "Path Home Classic", 0xf1f6f7, 0x26303a, 0x0e7063, 0xffffff, 0xedf6f3, 0x122a3e, 0xa5c9bf},
	{"aurora", "Aurora Pulse", 0xf0f5ff, 0x172344, 0x5331b8, 0xffffff, 0xd9faff, 0x263768, 0x9455ff},
	{"prismatic", "Prismatic Riot", 0xfff5f8, 0x3f203e, 0x962755, 0xffffff, 0xffe152, 0x412031, 0xff4f91},
	{"solar", "Solar Flare", 0xfff7ed, 0x44251c, 0xa43d13, 0xffffff, 0xffcb52, 0x41231c, 0xfa3f93},
	{"abyss", "Abyss Current", 0xeff8ff, 0x132b50, 0x1556a8, 0xffffff, 0xc2f7f6, 0x123552, 0x2f80ff},
	{"nebula", "Nebula Orchid", 0xf6f1ff, 0x33204d, 0x6332a5, 0xffffff, 0xf4d8fc, 0x392454, 0xb99aff},
	{"crimson", "Crimson Ember", 0xfff3ed, 0x42202c, 0xa32644, 0xffffff, 0xffcb70, 0x482228, 0xf43f5e},
	{"bio", "Bio-Luminescence", 0xf0fbf4, 0x173a30, 0x176449, 0xffffff, 0xbff5c0, 0x173a30, 0x14e6bd},
}
func paletteFor(id string) visualPalette {
	for _, p := range palettes { if p.ID == id { return p } }
	return palettes[0]
}
func applyTheme() {
	activePalette = paletteFor(input.Theme)
	next := proc(pCreateBrush, rgb(activePalette.Background))
	if next == 0 { return }
	old := brush
	brush = next
	// The original class brush remains valid until process exit.
	if old != 0 && old != classBrush { proc(pDeleteObject, old) }
	proc(pInvalidateRect, hwnd, 0, 1)
}
func ptr(s string) *uint16                              { return syscall.StringToUTF16Ptr(s) }
//go:uintptrescapes
func proc(p *syscall.LazyProc, vals ...uintptr) uintptr { v, _, _ := p.Call(vals...); return v }
func create(cls, title string, style uint32, x, y, w, h int, parent uintptr, id int) uintptr {
	v, _, _ := pCreateWindow.Call(0, uintptr(unsafe.Pointer(ptr(cls))), uintptr(unsafe.Pointer(ptr(title))), uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), proc(pGetModule, 0), 0)
	if v == 0 {
		log.Printf("child window creation failed: class=%s id=%d", cls, id)
	} else if parent != 0 {
		controls = append(controls, v)
	}
	return v
}
func text(x, y, w, h int, s string, big bool) uintptr {
	c := create("STATIC", s, WS_CHILD|WS_VISIBLE, x, y, w, h, hwnd, 0)
	f := font
	if big {
		f = largeFont
	}
	proc(pSendMessage, c, WM_SETFONT, f, 1)
	return c
}
func title(x, y, w int, s string) uintptr {
	c := text(x, y, w, 43, s, true)
	proc(pSendMessage, c, WM_SETFONT, titleFont, 1)
	return c
}
func btn(x, y, w, h int, s string, id int, fn func()) uintptr {
	// Record the caption before CreateWindowEx can synchronously request painting.
	buttonLabels[uint32(id)] = s
	c := create("BUTTON", s, WS_CHILD|WS_VISIBLE|WS_TABSTOP|0x000B, x, y, w, h, hwnd, id)
	proc(pSendMessage, c, WM_SETFONT, buttonFont, 1)
	clickActions[id] = fn
	return c
}
// Owner-drawn push buttons retain the native BUTTON click and focus behavior.
func paintButton(d *drawItem) {
	r := d.Rect
	proc(pFillRect, d.DC, uintptr(unsafe.Pointer(&r)), brush)
	// Avoid a synchronous window-message round trip inside WM_DRAWITEM.
	label := buttonLabels[d.CtlID]
	nav := d.CtlID >= 100 && d.CtlID < 111
	secondary := nav || d.CtlID == 900 || d.CtlID == 901 || d.CtlID == 903 || (d.CtlID >= 510 && d.CtlID <= 516)
	active := strings.HasPrefix(label, "• ")
	p := activePalette
	fill, edge, ink := rgb(p.Primary), rgb(p.Edge), rgb(p.PrimaryInk)
	if secondary && !active { fill, edge, ink = rgb(p.Secondary), rgb(p.Edge), rgb(p.SecondaryInk) }
	if d.ItemState&1 != 0 { edge = rgb(p.Ink) }
	if d.ItemState&4 != 0 { fill, edge, ink = 0x00e7e7e7, 0x00cccccc, 0x00666666 }
	b := proc(pCreateBrush, fill)
	pen := proc(pCreatePen, 0, 1, edge)
	oldB := proc(pSelectObject, d.DC, b)
	oldP := proc(pSelectObject, d.DC, pen)
	proc(pRoundRect, d.DC, uintptr(r.Left+1), uintptr(r.Top+1), uintptr(r.Right-1), uintptr(r.Bottom-1), 14, 14)
	proc(pSelectObject, d.DC, oldB)
	proc(pSelectObject, d.DC, oldP)
	proc(pDeleteObject, b)
	proc(pDeleteObject, pen)
	oldF := proc(pSelectObject, d.DC, buttonFont)
	proc(pSetBkMode, d.DC, 1)
	proc(pSetTextColor, d.DC, ink)
	tr := buttonRect{r.Left+10,r.Top+2,r.Right-10,r.Bottom-2}
	proc(pDrawText, d.DC, uintptr(unsafe.Pointer(ptr(label))), ^uintptr(0), uintptr(unsafe.Pointer(&tr)), 0x0001|0x0004|0x0020|0x0800|0x8000)
	proc(pSelectObject, d.DC, oldF)
	if d.ItemState&0x10 != 0 {
		fr := buttonRect{r.Left+4,r.Top+4,r.Right-4,r.Bottom-4}
		proc(pDrawFocusRect, d.DC, uintptr(unsafe.Pointer(&fr)))
	}
}
func edit(key string, x, y, w, h int, multiline bool) uintptr {
	style := uint32(WS_CHILD | WS_VISIBLE | WS_BORDER | WS_TABSTOP)
	if multiline {
		style |= ES_MULTILINE | ES_AUTOVSCROLL | WS_VSCROLL
	}
	// Stable IDs let runtime tests locate actual edit controls without relying on Z-order.
	ids := map[string]int{"state":1200, "county":1201, "zip":1202, "release":1203, "start":1204, "end":1205, "supervision":1206, "support":1207, "appointmentDate":1208, "appointmentText":1209, "cash":1210, "birthCost":1211, "idCost":1212, "otherCost":1213, "asked":1214, "unstuckGoal":1215}
	c := create("EDIT", fieldText(key), style, x, y, w, h, hwnd, ids[key])
	proc(pSendMessage, c, WM_SETFONT, font, 1)
	fields[key] = c
	return c
}
func fieldText(k string) string {
	switch k {
	case "state":
		return input.State
	case "county":
		return input.County
	case "zip":
		return input.Zip
	case "release":
		return input.Release
	case "start":
		return input.Start
	case "end":
		return input.End
	case "supervision":
		return input.Supervision
	case "support":
		return input.Support
	case "appointmentDate":
		return input.AppointmentDate
	case "appointmentText":
		return input.AppointmentText
	case "cash":
		if !budgetAmountKnown("cash", input.Cash) { return "" }
		return fmt.Sprintf("%.2f", input.Cash)
	case "birthCost":
		if !budgetAmountKnown("birthCost", input.BirthCost) { return "" }
		return fmt.Sprintf("%.2f", input.BirthCost)
	case "idCost":
		if !budgetAmountKnown("idCost", input.IDCost) { return "" }
		return fmt.Sprintf("%.2f", input.IDCost)
	case "otherCost":
		if !budgetAmountKnown("otherCost", input.OtherCost) { return "" }
		return fmt.Sprintf("%.2f", input.OtherCost)
	case "asked":
		return input.Asked
	case "unstuckGoal":
		return input.UnstuckGoal
	}
	return ""
}
func getText(h uintptr) string {
	if h == 0 {
		return ""
	}
	n := proc(pGetTextLength, h)
	b := make([]uint16, n+2)
	// GetWindowTextW(HWND, LPWSTR, int): buffer precedes capacity.
	proc(pGetText, h, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	return syscall.UTF16ToString(b)
}
func budgetAmountKnown(key string, amount float64) bool {
	if input.BudgetKnown != nil { if known, exists := input.BudgetKnown[key]; exists { return known } }
	return amount > 0 // Legacy zero cannot be distinguished from an unknown fee.
}
func money(v string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(v, "$", "")), 64)
	if n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	return n
}
func syncFields() {
	for key, c := range fields {
		v := strings.TrimSpace(getText(c))
		switch key {
		case "state":
			input.State = strings.ToUpper(v)
		case "county":
			input.County = v
		case "zip":
			input.Zip = v
		case "release":
			input.Release = v
		case "start":
			input.Start = v
		case "end":
			input.End = v
		case "supervision":
			input.Supervision = v
		case "support":
			input.Support = v
		case "appointmentDate":
			input.AppointmentDate = v
		case "appointmentText":
			input.AppointmentText = v
		case "cash":
			if input.BudgetKnown == nil { input.BudgetKnown = map[string]bool{} }
			input.BudgetKnown[key] = v != ""
			input.Cash = money(v)
		case "birthCost":
			if input.BudgetKnown == nil { input.BudgetKnown = map[string]bool{} }
			input.BudgetKnown[key] = v != ""
			input.BirthCost = money(v)
		case "idCost":
			if input.BudgetKnown == nil { input.BudgetKnown = map[string]bool{} }
			input.BudgetKnown[key] = v != ""
			input.IDCost = money(v)
		case "otherCost":
			if input.BudgetKnown == nil { input.BudgetKnown = map[string]bool{} }
			input.BudgetKnown[key] = v != ""
			input.OtherCost = money(v)
		case "asked":
			input.Asked = v
		case "unstuckGoal":
			input.UnstuckGoal = v
		}
	}
}

// The app uses a DPAPI-encrypted file scoped to the current Windows user.
// Legacy v0.8 plaintext profiles are migrated once only after verified encrypted write.
func storePath() string {
	dir, e := os.UserConfigDir()
	if e != nil {
		return ""
	}
	dir = filepath.Join(dir, "PathHome")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return ""
	}
	return filepath.Join(dir, "plan.dpapi")
}
func legacyStorePath() string {
	dir, e := os.UserConfigDir()
	if e != nil {
		return ""
	}
	return filepath.Join(dir, "ReentryNavigator", "plan.json")
}
func loadPrivatePlan() error {
	path := storePath()
	if path == "" {
		return fmt.Errorf("cannot locate secure user profile folder")
	}
	if encrypted, err := os.ReadFile(path); err == nil {
		raw, err := unprotectLocal(encrypted)
		if err != nil {
			return fmt.Errorf("encrypted profile could not be opened: %w", err)
		}
		defer clear(raw)
		if isVault(raw) {
			if _, err := vaultSalt(raw); err != nil { return err }
			vaultEnvelope = bytes.Clone(raw)
			vaultLocked = true
			return nil
		}
		return json.Unmarshal(raw, &input)
	} else if !os.IsNotExist(err) {
		return err
	}
	old := legacyStorePath()
	if old == "" {
		return nil
	}
	if raw, err := os.ReadFile(old); err == nil {
		var previous state
		if err := json.Unmarshal(raw, &previous); err != nil {
			return fmt.Errorf("legacy profile could not be read: %w", err)
		}
		input = previous
		if err := saveEncrypted(raw); err != nil {
			return fmt.Errorf("legacy profile was NOT deleted because secure migration failed: %w", err)
		}
		if err := os.Remove(old); err != nil {
			return fmt.Errorf("encrypted migration succeeded, but old plaintext profile still needs secure removal: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}
func saveEncrypted(raw []byte) error {
	path := storePath()
	if path == "" {
		return fmt.Errorf("secure user directory unavailable")
	}
	if vaultLocked { return fmt.Errorf("private vault is locked") }
	payload := raw
	if len(vaultEnvelope)>0 {
		salt,err:=vaultSalt(vaultEnvelope)
		if err!=nil{return err}
		payload,err=sealVault(vaultKey,salt,raw)
		if err!=nil{return err}
	}
	encrypted, err := protectLocal(payload)
	if err != nil { return err }
	if err=atomicPrivateWrite(path, encrypted);err!=nil{return err}
	if len(vaultEnvelope)>0 {vaultEnvelope=payload}
	return nil
}
func save(silent bool) bool { return saveWithFeedback(silent,true) }
func saveWithFeedback(silent,showErrors bool) bool {
	if vaultLocked {return false}
	fail:=func(msg string){if showErrors{alert(msg)}}
	// Validate the live entry fields before syncFields can coerce bad numeric text to zero.
	for _, k := range []string{"cash", "birthCost", "idCost", "otherCost"} {
		if c, ok := fields[k]; ok {
			if err := validateMoneyField(getText(c)); err != nil {
				fail("Check the amount for " + k + ": " + err.Error() + ". Leave blank if unknown.")
				return false
			}
		}
	}
	// Restore the previously accepted state if validation or encrypted write fails.
	previous := input
	if input.BudgetKnown != nil {
		previous.BudgetKnown = make(map[string]bool, len(input.BudgetKnown))
		for k, v := range input.BudgetKnown { previous.BudgetKnown[k] = v }
	}
	committed := false
	defer func() { if !committed { input = previous } }()
	syncFields()
	if !knownJurisdiction(input.State, data.Jurisdictions) {
		fail("Use a valid two-letter state abbreviation for where you live (for example MI), or leave blank. Enter federal supervised release in the supervision field.")
		return false
	}
	if invalidDate(input.Release) || invalidDate(input.Start) || invalidDate(input.End) || invalidDate(input.AppointmentDate) {
		fail("Enter dates as YYYY-MM-DD, or leave them blank if unknown.")
		return false
	}
	if !dateOrderValid(input.Start, input.End) {
		fail("The expected supervision end date cannot be before the supervision start date. Confirm both dates.")
		return false
	}
	b, err := json.MarshalIndent(input, "", "  ")
	if err == nil {
		err = saveEncrypted(b)
	}
	clear(b)
	if err != nil {
		fail("Could not securely save your plan: " + err.Error())
		return false
	}
	committed = true
	if !silent && showErrors {
		alert("Saved securely on this device. This is not a synced account.")
	}
	return true
}

func invalidDate(v string) bool {
	if v == "" {
		return false
	}
	t, e := time.Parse("2006-01-02", v)
	return e != nil || t.Format("2006-01-02") != v
}
func alert(msg string) {
	proc(pMessageBox, hwnd, uintptr(unsafe.Pointer(ptr(msg))), uintptr(unsafe.Pointer(ptr("Path Home"))), 0)
}
func clearPanel() {
	// State is committed explicitly by actions; redrawing must not silently accept invalid fields.
	for _, c := range controls {
		if c != 0 {
			proc(pDestroyWindow, c)
		}
	}
	controls = nil
	releaseBitmaps()
	buttonLabels = map[uint32]string{}
	fields = map[string]uintptr{}
	clickActions = map[int]func(){}
	lastLink = ""
}
func releaseBitmaps() {
	for _, h := range bitmaps {
		proc(pDeleteObject, h)
	}
	bitmaps = nil
}
func photo(asset string, x, y int) {
	p := filepath.Join(assetDir, "photos", asset+".bmp")
	h := proc(pLoadImage, 0, uintptr(unsafe.Pointer(ptr(p))), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE)
	if h == 0 {
		return
	}
	width, height := 126, 82
	if asset == "hero" {
		width, height = 232, 128
	}
	c := create("STATIC", "", WS_CHILD|WS_VISIBLE|SS_BITMAP, x, y, width, height, hwnd, 0)
	if c == 0 {
		proc(pDeleteObject, h)
		return
	}
	proc(pSendMessage, c, STM_SETIMAGE, IMAGE_BITMAP, h)
	bitmaps = append(bitmaps, h)
	// A static control may make its own bitmap copy; track that handle too.
	if actual := proc(pSendMessage, c, 0x0173, IMAGE_BITMAP, 0); actual != 0 && actual != h {
		bitmaps = append(bitmaps, actual)
	}
}
func navGlyph(asset string, x, y, size int) {
	path := filepath.Join(assetDir, "icons", asset+".bmp")
	h := proc(pLoadImage, 0, uintptr(unsafe.Pointer(ptr(path))), IMAGE_BITMAP, 0, 0, LR_LOADFROMFILE)
	if h == 0 {
		return
	}
	c := create("STATIC", "", WS_CHILD|WS_VISIBLE|SS_BITMAP, x, y, size, size, hwnd, 0)
	if c == 0 {
		proc(pDeleteObject, h)
		return
	}
	proc(pSendMessage, c, STM_SETIMAGE, IMAGE_BITMAP, h)
	bitmaps = append(bitmaps, h)
	// A static control may make its own bitmap copy; track that handle too.
	if actual := proc(pSendMessage, c, 0x0173, IMAGE_BITMAP, 0); actual != 0 && actual != h {
		bitmaps = append(bitmaps, actual)
	}
}
func nav(target string) {
	if !save(true) {
		return
	}
	page = target
	draw()
}
func strWrap(s string, n int) []string {
	words := strings.Fields(s)
	var out []string
	row := ""
	for _, w := range words {
		if len(row)+len(w) > n && row != "" {
			out = append(out, row)
			row = ""
		}
		if row != "" {
			row += " "
		}
		row += w
	}
	if row != "" {
		out = append(out, row)
	}
	return out
}
func paragraph(x, y, w int, s string, maxRows int) int {
	lines := strWrap(s, max(25, w/9))
	if maxRows > 0 && len(lines) > maxRows {
		lines = lines[:maxRows]
	}
	for _, ln := range lines {
		text(x, y, w, 24, ln, false)
		y += 27
	}
	return y
}
func link(url string) {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return
	}
	proc(pShellExec, hwnd, uintptr(unsafe.Pointer(ptr("open"))), uintptr(unsafe.Pointer(ptr(url))), 0, 0, 1)
}
func addNav(x, y int) {
	navs := []struct{ Page, Name string }{{"home", "Home / Start Here"}, {"plan", "My Release Plan"}, {"ask", "Am I Allowed To...?"}, {"library", "Law Library"}, {"appointments", "Appointments"}, {"money", "Money & Documents"}, {"help", "Find Help Nearby"}, {"family", "Family Preparation"}, {"progress", "My Progress"}, {"unstuck", "Get Unstuck"}, {"stay", "Stay on Track"}}
	for i, n := range navs {
		n := n
		glyph := map[string]string{"home": "home", "plan": "plan", "ask": "ask", "library": "law", "appointments": "appointments", "money": "money", "help": "help", "family": "family", "progress": "progress", "unstuck": "help", "stay": "stay"}[n.Page]
		navGlyph(glyph, x+3, y+i*51+8, 32)
		label := n.Name
		if page == n.Page {
			label = "• " + label
		}
		btn(x+40, y+i*51, 196, 45, label, 100+i, func() { nav(n.Page) })
	}
}
func draw() {
	if vaultLocked {drawVaultUnlock();return}
	clearPanel()
	navGlyph("brand", 35, 13, 72)
	title(115, 19, 750, "PATH HOME · STRUCTURAL REENTRY ASSISTANCE")
	text(115, 64, 750, 30, "Your next step, made clear.    |    Not a government service · Private local tools", false)
	addNav(22, 128)
	btn(62, 690, 196, 42, "Privacy & lock", 904, func() { nav("privacy") })
	btn(883, 22, 194, 45, "Our Mission", 900, func() { nav("mission") })
	btn(883, 74, 194, 34, "Appearance", 903, func() { nav("appearance") })
	x := 284
	w := 788
	switch page {
	case "home":
		drawHome(x, w)
	case "mission":
		drawMission(x, w)
	case "appearance":
		drawAppearance(x, w)
	case "privacy":
		drawPrivacy(x,w)
	case "plan":
		drawPlan(x, w)
	case "ask":
		drawAsk(x, w)
	case "appointments":
		drawAppointments(x, w)
	case "money":
		drawMoney(x, w)
	case "help":
		drawHelp(x, w)
	case "family":
		drawFamily(x, w)
	case "progress":
		drawProgress(x, w)
	case "unstuck":
		drawUnstuck(x, w)
	case "stay":
		drawStay(x, w)
	case "library":
		drawLibrary(x, w)
	}
	text(286, 744, 790, 25, "General information only. Check signed release conditions and official agency instructions.", false)
	text(286, 773, 650, 22, "Published by Genyxis Network", false)
}

func drawAppearance(x, w int) {
	title(x, 118, w, "Make Path Home yours")
	paragraph(x, 172, 745, "Choose a visual palette. Your selection stays on this device. The original logo, mission and navigation remain unchanged.", 3)
	for i, p := range palettes {
		p := p
		label := p.Name
		if p.ID == activePalette.ID { label = "• " + label }
		btn(x+(i%2)*380, 275+(i/2)*82, 355, 64, label, 910+i, func() {
			previous := input.Theme
			input.Theme = p.ID
			if !save(true) { input.Theme = previous; return }
			applyTheme()
			draw()
		})
	}
	text(x, 625, 740, 30, "Current palette: "+activePalette.Name, true)
	btn(x, 671, 250, 45, "Return Home", 920, func() { nav("home") })
}
func drawMission(x, w int) {
	title(x, 115, w, "Our Mission")
	text(x, 168, 760, 28, "Because freedom deserves a fighting chance.", true)
	raw, err := bundled.ReadFile("assets/mission.txt")
	if err != nil {
		paragraph(x, 238, 760, "Mission statement unavailable; please check the application files.", 3)
		return
	}
	display := strings.ReplaceAll(strings.TrimSpace(string(raw)), "\n", "\r\n")
	viewer := create("EDIT", display, WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_READONLY|ES_AUTOVSCROLL, x, 215, 758, 486, hwnd, 902)
	proc(pSendMessage, viewer, WM_SETFONT, font, 1)
}

func drawHome(x, w int) {
	title(x, 116, 555, "Welcome home. Start here.")
	text(x, 166, 545, 25, "Big steps begin with small ones. Choose what you need.", false)
	text(x, 203, 555, 32, "Because freedom deserves a fighting chance.", false)
	photo("hero", x+555, 115)
	choices := []struct{ Lbl, Desc, P, Img string }{
		{"My release plan", "Dates, reporting, first steps", "plan", "release_plan"},
		{"Am I allowed to...?", "Travel, housing, jobs, voting", "ask", "appointments"},
		{"Money & ID papers", "Birth certificate, Social Security, ID", "money", "money"},
		{"Find help near me", "Housing, transportation, work", "help", "local_help"},
		{"Appointments", "Keep deadlines and check-ins", "appointments", "appointments"},
		{"Family preparation", "Get a loved one ready", "family", "family"},
	}
	for i, c := range choices {
		c := c
		col := i % 2
		row := i / 2
		xx := x + col*389
		yy := 258 + row*123
		photo(c.Img, xx, yy)
		btn(xx+133, yy+1, 242, 55, c.Lbl, 200+i, func() { nav(c.P) })
		paragraph(xx+134, yy+59, 236, c.Desc, 2)
	}
	btn(x+545, 646, 225, 43, "Stuck? Find a way forward", 430, func() { nav("unstuck") })
	btn(x+545, 692, 225, 33, "Read Our Mission", 901, func() { nav("mission") })
	title(x, 635, 530, "My supervision timeline")
	paragraph(x, 683, 520, timelineLabel(), 2)
}

func timelineLabel() string {
	if input.End == "" {
		return "Add your expected supervision end date to see time remaining."
	}
	t, err := time.Parse("2006-01-02", input.End)
	if err != nil {
		return "Check your date format."
	}
	days := int(time.Until(t).Hours() / 24)
	if days < 0 {
		return "Expected end date has passed. Confirm official discharge status."
	}
	return fmt.Sprintf("Expected end: %s  |  %d days remaining  |  verify with supervising agency", t.Format("Jan 2, 2006"), days)
}
func drawPlan(x, w int) {
	title(x, 118, w, "Step 1 — Tell us about your release")
	text(x, 169, 760, 29, "You can leave unknown dates blank and add them later.", false)
	labels := []struct{ k, label string }{{"state", "State abbreviation (for example MI, CA, NY)"}, {"county", "County"}, {"zip", "ZIP code"}, {"supervision", "Type of supervision (parole, federal, probation, none)"}, {"release", "Release date (YYYY-MM-DD)"}, {"start", "Supervision start date (YYYY-MM-DD)"}, {"end", "Expected supervision end date (YYYY-MM-DD)"}, {"support", "How long do you want help? (90 days, ongoing, through supervision)"}}
	for i, l := range labels {
		c := i % 2
		r := i / 2
		xx := x + c*385
		yy := 224 + r*102
		text(xx, yy, 360, 40, l.label, false)
		edit(l.k, xx, yy+42, 348, 38, false)
	}
	btn(x, 655, 280, 55, "Save my dates and plan", 350, func() {
		if save(false) {
			draw()
		}
	})
	btn(x+305, 655, 240, 55, "Next: Appointments", 351, func() { nav("appointments") })
}
func drawAsk(x, w int) {
	title(x, 120, w, "Am I Allowed To...?")
	text(x, 170, 770, 29, "Type any situation. We show relevant sources, not automatic permission.", false)
	text(x, 219, 700, 27, "Your question or situation", false)
	edit("asked", x, 250, 730, 96, true)
	btn(x, 363, 275, 52, "Find relevant legal information", 360, func() {
		if save(true) {
			draw()
		}
	})
	text(x, 432, 735, 30, "Common questions: Can I travel? Move? Get a job? Vote?", false)
	text(x, 469, 735, 30, "Can I live with family? Drive? Leave the county?", false)
	if input.Asked != "" {
		drawAskResults(x, w)
	} else {
		paragraph(x, 536, 745, "Only your signed conditions and the supervising agency can confirm your individual requirements. If legal sources disagree or are incomplete, ask your attorney or officer.", 5)
	}
}
func drawAskResults(x, w int) {
	stateCode := strings.ToUpper(strings.TrimSpace(input.State))
	query := strings.ToLower(input.Asked)
	keyword := questionTopic(query)
	var results []legalRule
	for _, r := range data.Rules {
		if r.Jurisdiction != stateCode && !(federalSupervision(input.Supervision) && r.Jurisdiction == "FED") {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(r.Topic+" "+r.Title), keyword) {
			continue
		}
		results = append(results, r)
	}
	title(x, 537, 720, fmt.Sprintf("Available research — %s", stateCode))
	if stateCode == "" {
		paragraph(x, 581, 750, "Choose your state in My Release Plan before relying on jurisdiction-specific legal research. The Law Library remains available.", 4)
		return
	}
	if keyword == "multiple" {
		paragraph(x, 585, 750, "Your question may involve several legal rules. Ask separately about each issue (for example, residence approval AND travel permission), and confirm with your supervising agency or an attorney.", 4)
		return
	}
	if keyword == "" || len(results) == 0 {
		paragraph(x, 585, 750, "No matching researched rule for that particular question. Do not assume permission. Use the Law Library or ask an attorney or supervising agency.", 4)
		return
	}
	for i := 0; i < 2 && i < len(results); i++ {
		r := results[i]
		yy := 585 + i*67
		text(x, yy, 750, 24, r.Title+"  [review needed]", false)
		text(x, yy+27, 745, 35, firstN(r.Plain, 100), false)
		if i == 0 {
			lastLink = r.Source
		}
	}
	if lastLink != "" {
		btn(x+570, 362, 190, 52, "Open official source", 361, func() { link(lastLink) })
	}
}
func firstN(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}
func drawAppointments(x, w int) {
	title(x, 123, w, "My appointments & check-ins")
	paragraph(x, 177, 755, "Enter the exact reporting deadline written in your release paperwork. This app saves your note locally; it does not send automatic alerts or contact your officer.", 3)
	text(x, 276, 700, 30, "Appointment date (YYYY-MM-DD)", false)
	edit("appointmentDate", x, 308, 340, 43, false)
	text(x, 378, 700, 30, "What is this appointment? Where do you need to go?", false)
	edit("appointmentText", x, 411, 720, 111, true)
	btn(x, 554, 290, 55, "Save appointment", 375, func() { save(false) })
	btn(x+315, 554, 250, 55, "Check my dates", 376, func() { if save(true) { draw() } })
	title(x, 633, 720, "Your next appointment")
	appt := input.AppointmentDate
	if appt == "" {
		appt = "Not entered yet"
	}
	text(x, 681, 760, 34, appt+"   "+firstN(input.AppointmentText, 50), false)
}
func drawMoney(x, w int) {
	title(x, 122, w, "Money for your important papers")
	paragraph(x, 171, 745, "A replacement Social Security card is free. Birth certificate and state ID fees vary; ask about fee waivers. You can edit these estimates.", 3)
	labels := []struct{ k, label string }{{"cash", "Money available ($)"}, {"birthCost", "Birth certificate estimate ($)"}, {"idCost", "State ID estimate ($)"}, {"otherCost", "Travel, postage and other costs ($)"}}
	for i, l := range labels {
		c := i % 2
		r := i / 2
		xx := x + c*375
		yy := 280 + r*104
		text(xx, yy, 340, 28, l.label, false)
		edit(l.k, xx, yy+31, 318, 41, false)
	}
	need := input.BirthCost + input.IDCost + input.OtherCost - input.Cash
	if need < 0 {
		need = 0
	}
	complete := budgetAmountKnown("cash", input.Cash) && budgetAmountKnown("birthCost", input.BirthCost) && budgetAmountKnown("idCost", input.IDCost) && budgetAmountKnown("otherCost", input.OtherCost)
	if complete {
		title(x, 515, 745, fmt.Sprintf("Estimated still needed: $%.2f", need))
	} else if !budgetAmountKnown("cash", input.Cash) {
		title(x, 515, 745, "Partial estimate — available money unknown")
	} else {
		title(x, 515, 745, fmt.Sprintf("Partial shortfall (fees missing): $%.2f", need))
	}
	paragraph(x, 565, 740, "Leave unknown amounts blank. Enter zero only for a known zero amount. Confirm fees and waivers.", 2)
	btn(x, 620, 260, 56, "Update my budget", 380, func() {
		if save(true) {
			draw()
		}
	})
	btn(x+285, 620, 245, 56, "Find funding assistance", 381, func() { nav("help") })
}
func drawHelp(x, w int) {
	title(x, 117, w, "Find real help in my county")
	stateCode := strings.ToUpper(strings.TrimSpace(input.State))
	county := strings.TrimSpace(input.County)
	if stateCode == "" || county == "" {
		paragraph(x, 175, 740, "Choose your state and county in My Release Plan first. Path Home will not guess your location or show an unrelated office.", 4)
		btn(x, 280, 320, 55, "Choose my state and county", 501, func() { nav("plan") })
		btn(x, 355, 250, 55, "Open local 211 finder", 502, func() { link("https://211.org/") })
		return
	}
	text(x, 167, 740, 29, "Showing verified directory contacts for "+county+", "+stateCode, false)
	text(x, 196, 750, 24, "Locations selected by you. No device location sent to Genyxis.", false)
	needs := []string{"All", "documents", "housing", "employment", "transportation", "legal", "money"}
	for i, category := range needs {
		c := category
		label := category
		if category == input.HelpNeed || (input.HelpNeed == "" && category == "All") {
			label = "• " + category
		}
		col := i % 4
		row := i / 4
		btn(x+col*190, 239+row*48, 181, 43, label, 510+i, func() { oldNeed, oldPage := input.HelpNeed, input.HelpPage; input.HelpNeed = c; input.HelpPage = 0; if save(true) { draw() } else { input.HelpNeed, input.HelpPage = oldNeed, oldPage } })
	}
	matches := findLocalProviders(data.LocalProviders, stateCode, county, input.HelpNeed)
	if len(matches) == 0 {
		paragraph(x, 356, 740, "No source-checked county match for this category yet. Do not travel to a guessed office. Call 211 or use its directory to find the right service area, then confirm eligibility and hours.", 5)
	} else {
		pageCount := (len(matches) + 2) / 3
		if input.HelpPage >= pageCount || input.HelpPage < 0 {
			input.HelpPage = 0
		}
		text(x, 351, 755, 27, fmt.Sprintf("%d local contacts · page %d of %d", len(matches), input.HelpPage+1, pageCount), true)
		start := input.HelpPage * 3
		for i := start; i < len(matches) && i < start+3; i++ {
			r := matches[i]
			yy := 390 + (i-start)*89
			text(x, yy, 552, 24, firstN(r.Name, 67), true)
			text(x, yy+24, 560, 24, "Phone: "+r.Phone+"  |  "+firstN(r.Scope, 72), false)
			text(x, yy+48, 748, 23, "Before you go: "+firstN(r.BeforeYouGo, 91), false)
			btn(x+573, yy, 165, 39, "Official details", 530+i-start, func() { link(r.Website) })
		}
		if input.HelpPage > 0 {
			btn(x, 666, 170, 43, "Previous", 549, func() { old := input.HelpPage; input.HelpPage--; if save(true) { draw() } else { input.HelpPage = old } })
		}
		if input.HelpPage+1 < pageCount {
			btn(x+185, 666, 170, 43, "More local help", 550, func() { old := input.HelpPage; input.HelpPage++; if save(true) { draw() } else { input.HelpPage = old } })
		}
	}
	btn(x+368, 666, 195, 43, "Get unstuck", 551, func() { nav("unstuck") })
	btn(x+573, 666, 165, 43, "211 backup", 552, func() { link("https://211.org/") })
	text(x, 713, 740, 23, "Call before visiting. Funding, eligibility and appointments must be confirmed.", false)
}
func drawFamily(x, w int) {
	title(x, 120, w, "Help someone come home")
	paragraph(x, 175, 740, "Family and friends can prepare documents, a ride, housing, and a first-week plan. This offline version stays on this computer—there is no live invitation or shared account yet.", 4)
	tasks := []string{"Confirm release day and pickup plan", "Find a safe place to stay", "Make a birth certificate plan", "Check Social Security replacement process", "Help with state ID or license", "Prepare a phone or communication plan", "List reporting appointments", "Find transportation options", "Check food and healthcare support", "Look up jobs and training"}
	for i, t := range tasks {
		xx := x + (i%2)*385
		yy := 300 + (i/2)*65
		id := 600 + i
		c := create("BUTTON", t, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX|0x2000, xx, yy, 368, 51, hwnd, id)
		proc(pSendMessage, c, WM_SETFONT, font, 1)
		if input.Checklist[t] {
			proc(pSendMessage, c, BM_SETCHECK, 1, 0)
		}
		t := t
		clickActions[id] = func() { old := input.Checklist[t]; input.Checklist[t] = proc(pSendMessage, c, BM_GETCHECK, 0, 0) != 0; if !save(true) { input.Checklist[t] = old; checked := uintptr(0); if old { checked = 1 }; proc(pSendMessage, c, BM_SETCHECK, checked, 0) } }
	}
	btn(x, 650, 270, 53, "Save preparation checklist", 400, func() { save(false) })
	btn(x+293, 650, 310, 53, "Review release plan dates", 401, func() { nav("plan") })
}
func drawProgress(x, w int) {
	title(x, 118, w, "My progress and next step")
	text(x, 177, 760, 28, timelineLabel(), false)
	tasks := []string{"Confirm actual first reporting appointment", "Get or order birth certificate", "Check Social Security replacement eligibility", "Apply for an ID or license", "Confirm stable housing", "Make a first-week transportation plan", "Contact one employment resource", "Check food and healthcare options"}
	done := 0
	for _, k := range tasks {
		if input.Checklist[k] {
			done++
		}
	}
	title(x, 230, 760, fmt.Sprintf("%d of %d first steps completed", done, len(tasks)))
	for i, t := range tasks {
		y := 300 + i*48
		id := 700 + i
		c := create("BUTTON", t, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, x, y, 755, 42, hwnd, id)
		proc(pSendMessage, c, WM_SETFONT, font, 1)
		if input.Checklist[t] {
			proc(pSendMessage, c, BM_SETCHECK, 1, 0)
		}
		t := t
		clickActions[id] = func() { old := input.Checklist[t]; input.Checklist[t] = proc(pSendMessage, c, BM_GETCHECK, 0, 0) != 0; if save(true) { draw() } else { input.Checklist[t] = old; checked := uintptr(0); if old { checked = 1 }; proc(pSendMessage, c, BM_SETCHECK, checked, 0) } }
	}
	btn(x, 684, 235, 38, "Save my progress", 410, func() { save(false) })
	btn(x+251, 684, 370, 38, "A step is blocked? Find another way", 411, func() { nav("unstuck") })
}

// A simple, voluntary Lean-style barrier-removal workflow. No scores or identifying data leave this device.
func drawUnstuck(x, w int) {
	title(x, 118, w, "I'm stuck. Help me find another way.")
	paragraph(x, 166, 750, "Tell us what is in your way. You are not the problem. This tool finds a practical next step, not a grade or a parole decision.", 3)
	text(x, 251, 755, 31, "What are you trying to get done? (optional)", true)
	edit("unstuckGoal", x, 291, 748, 40, false)
	text(x, 339, 750, 27, "What is getting in the way?", true)
	names := []string{"Money or fees", "Missing ID or paperwork", "No transportation", "Program said no", "Instructions unclear", "Need safe housing", "Something else"}
	for i, label := range names {
		label := label
		row, col := i/2, i%2
		yy := 374 + row*48
		btn(x+col*382, yy, 367, 42, label, 800+i, func() {
			old := input.UnstuckBarrier
			input.UnstuckBarrier = label
			if save(true) { draw() } else { input.UnstuckBarrier = old }
		})
	}
	selected := input.UnstuckBarrier
	if selected == "" {
		selected = "Choose a situation above"
	}
	text(x, 568, 744, 27, "Selected: "+selected, false)
	paragraph(x, 601, 740, barrierNextAction(input.UnstuckBarrier), 3)
	btn(x, 686, 310, 37, "Find the right local help", 815, func() { oldNeed, oldPage := input.HelpNeed, input.HelpPage; input.HelpNeed = barrierHelpCategory(input.UnstuckBarrier); input.HelpPage = 0; if save(true) { page = "help"; draw() } else { input.HelpNeed, input.HelpPage = oldNeed, oldPage } })
	btn(x+331, 686, 306, 37, "Review my next steps", 816, func() { nav("progress") })
}

func drawLibrary(x, w int) {
	title(x, 119, w, "Law Library — free case law")
	paragraph(x, 175, 755, "Find court opinions and laws without a paid subscription. Old or overturned decisions may not control your case.", 3)
	btn(x, 266, 360, 54, "Open CourtListener (free cases)", 450, func() { link("https://www.courtlistener.com/") })
	btn(x, 335, 360, 54, "Open Google Scholar (case law)", 451, func() { link("https://scholar.google.com/") })
	btn(x, 404, 460, 54, "Library of Congress: free case-law guide", 452, func() { link("https://guides.loc.gov/free-case-law") })
	title(x, 488, w, "State-specific legal research")
	stateCode := strings.ToUpper(strings.TrimSpace(input.State))
	n := 0
	for _, r := range data.Rules {
		if r.Jurisdiction != stateCode || !strings.HasPrefix(r.Source, "https://") {
			continue
		}
		rr := r
		yy := 544 + n*51
		text(x, yy, 580, 33, firstN(rr.Title, 60), false)
		btn(x+590, yy, 160, 40, "Open source", 460+n, func() { link(rr.Source) })
		n++
		if n == 3 {
			break
		}
	}
	if n == 0 {
		paragraph(x, 550, 745, "Choose your state in My Release Plan. Local coverage is still limited; these are research links, not case-specific legal advice.", 3)
	}
}
func drawStay(x, w int) {
	title(x, 122, w, "Stay on Track")
	paragraph(x, 175, 755, "Your private moment to think about the direction you're going. No reports are sent to Path Home, family, or supervision agencies.", 3)
	if input.StayEnabled {
		text(x, 274, 755, 33, "Optional reminder: ON", true)
	} else {
		text(x, 274, 755, 33, "Optional reminder: OFF", true)
	}
	btn(x, 330, 320, 53, "Turn reminders on / off", 431, func() { old := input.StayEnabled; input.StayEnabled = !old; if save(true) { draw() } else { input.StayEnabled = old } })
	text(x, 417, 755, 36, "Are you in the right place, and the right time?", true)
	text(x, 456, 755, 30, "Just checking on you!", true)
	paragraph(x, 511, 755, "You've worked hard to get this far. Take a moment to think about where you're going and what matters most to you.", 3)
	btn(x, 608, 285, 53, "Show my private check-in", 432, func() {
		if input.StayEnabled {
			alert("Are you in the right place, and the right time? Just checking on you!\n\nYou've worked hard to get this far. Take a moment to think about where you're going and what matters most to you.")
		} else {
			alert("Turn on private reminders first.")
		}
	})
	paragraph(x, 686, 755, "Windows edition: manual check-ins only; no PC location sensor access, background location checks or location history.", 2)
}
func wndproc(h uintptr, msg uint32, w, l uintptr) uintptr {
	switch msg {
	case 0x002B: // WM_DRAWITEM
		if l != 0 {
			// WM_DRAWITEM supplies a valid DRAWITEMSTRUCT for this synchronous message.
			// Copy its bytes through Win32 rather than converting an integer to a Go pointer.
			var d drawItem
			proc(pCopyMemory, uintptr(unsafe.Pointer(&d)), l, unsafe.Sizeof(d))
			if d.CtlType == 4 {
				paintButton(&d)
				return 1
			}
		}
	case WM_COMMAND:
		// Only button-click notifications (BN_CLICKED) invoke these actions.
		if l == 0 || (w>>16)&0xffff != 0 {
			break
		}
		id := int(w & 0xffff)
		if fn, ok := clickActions[id]; ok {
			fn()
			return 0
		}
	case WM_CTLCOLORSTATIC:
		// Read-only EDIT uses WM_CTLCOLORSTATIC. Opaque text with matching brush prevents scroll ghosting.
		proc(pSetBkMode, w, 2)
		proc(pSetBkColor, w, rgb(activePalette.Background))
		proc(pSetTextColor, w, rgb(activePalette.Ink))
		return brush
	case 0x0014: // WM_ERASEBKGND: theme-aware canvas rather than stale class brush.
		var r buttonRect
		if proc(user32.NewProc("GetClientRect"), h, uintptr(unsafe.Pointer(&r))) != 0 {
			proc(pFillRect, w, uintptr(unsafe.Pointer(&r)), brush)
			return 1
		}
	case 0x0024: // WM_GETMINMAXINFO; fixed-coordinate layout must not be shrunk below its canvas.
		if l != 0 {
			var mm [5]point
			proc(pCopyMemory, uintptr(unsafe.Pointer(&mm[0])), l, unsafe.Sizeof(mm))
			mm[3] = point{1130, 850}
			proc(pCopyMemory, l, uintptr(unsafe.Pointer(&mm[0])), unsafe.Sizeof(mm))
			return 0
		}
	case 0x0113: // WM_TIMER
		if !vaultLocked && len(vaultKey)>0 && (vaultPendingLock || time.Since(vaultLastActivity)>=vaultIdle) {lockPrivateVault(true)}
		return 0
	case 0x02B1: // WM_WTSSESSION_CHANGE
		if w==7 && len(vaultKey)>0 {vaultPendingLock=true;lockPrivateVault(true)}
		return 0
	case WM_CLOSE:
		if !vaultLocked && !save(true) {
			return 0
		}
		proc(pDestroyWindow, h)
		return 0
	case WM_DESTROY:
		proc(pPostQuit, 0)
		return 0
	}
	return proc(pDefWindowProc, h, uintptr(msg), w, l)
}
func prepareAssets() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil { return "", err }
	dir = filepath.Join(dir, "PathHome", "art")
	for _, group := range []string{"photos", "icons"} {
		if err := os.MkdirAll(filepath.Join(dir, group), 0700); err != nil { return "", err }
		entries, err := bundled.ReadDir("assets/" + group)
		if err != nil { return "", err }
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bmp") { continue }
			b, err := bundled.ReadFile("assets/" + group + "/" + entry.Name())
			if err != nil { return "", err }
			dest := filepath.Join(dir, group, entry.Name())
			// Reuse verified bytes; repair only derived cache files, not private profiles.
			if current, err := os.ReadFile(dest); err == nil && bytes.Equal(current, b) { continue }
			if err := os.WriteFile(dest, b, 0600); err != nil { return "", err }
		}
	}
	ico, err := bundled.ReadFile("assets/PathHome.ico")
	if err != nil { return "", err }
	if err := os.WriteFile(filepath.Join(dir, "PathHome.ico"), ico, 0600); err != nil { return "", err }
	return dir, nil
}
func main() {
	// Win32 windows and their message queue belong to the creating OS thread.
	runtime.LockOSThread()
	os.Exit(run())
}
func run() (exitCode int) {
	// Local diagnostic metadata only; never log plan text or user input.
	if dir, err := os.UserCacheDir(); err == nil {
		dir = filepath.Join(dir, "PathHome")
		if os.MkdirAll(dir, 0700) == nil {
			if f, err := os.OpenFile(filepath.Join(dir, "startup.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
				defer f.Close()
				log.SetOutput(f)
			}
		}
	}
	defer func() {
		if p := recover(); p != nil {
			log.Printf("fatal runtime failure: type=%T", p)
			exitCode = 1
		}
		log.Printf("shutdown exit=%d", exitCode)
	}()
	log.Print("startup: loading bundled reference data")
	b, err := bundled.ReadFile("assets/reentry-public.json")
	if err != nil || json.Unmarshal(b, &data) != nil {
		log.Print("startup: bundled reference data invalid")
		alert("Path Home reference data could not be loaded. Please use a verified application build.")
		return 1
	}
	lease,err:=acquireProfileLease()
	if err!=nil {alert("Path Home cannot acquire its private profile. Close any other instance and try again. No data was changed.");return 1}
	defer syscall.CloseHandle(lease)
	defer func(){clear(vaultKey)}()
	log.Print("startup: opening local profile")
	if err := loadPrivatePlan(); err != nil {
		log.Print("startup: secure profile could not be opened")
		alert("Your saved plan needs attention: " + err.Error() + "\nExisting data has not been overwritten.")
		return 1
	}
	if input.Checklist == nil {
		input.Checklist = map[string]bool{}
	}
	log.Print("startup: preparing local art")
	assetDir, err = prepareAssets()
	if err != nil {
		log.Print("startup: asset preparation failed")
		alert("Path Home could not prepare its local art. Your private plan was not deleted.")
		return 1
	}
	inst := proc(pGetModule, 0)
	activePalette = paletteFor(input.Theme)
	bg := proc(pCreateBrush, rgb(activePalette.Background))
	brush = bg
	classBrush = bg
	cName := ptr("ReentryNavigatorNativeDesktop")
	iconPath := ptr(filepath.Join(assetDir, "PathHome.ico"))
	bigIcon := proc(pLoadImage, 0, uintptr(unsafe.Pointer(iconPath)), IMAGE_ICON, 64, 64, LR_LOADFROMFILE)
	smallIcon := proc(pLoadImage, 0, uintptr(unsafe.Pointer(iconPath)), IMAGE_ICON, 16, 16, LR_LOADFROMFILE)
	wc := wndClass{CbSize: uint32(unsafe.Sizeof(wndClass{})), WndProc: syscall.NewCallback(wndproc), Instance: inst, BgBrush: bg, ClassName: cName, Icon: bigIcon, SmallIcon: smallIcon}
	if proc(pRegisterClass, uintptr(unsafe.Pointer(&wc))) == 0 {
		log.Print("startup: window class registration failed")
		alert("Could not register Windows app window.")
		return 1
	}
	log.Print("startup: creating main window")
	h, _, createErr := pCreateWindow.Call(0, uintptr(unsafe.Pointer(cName)), uintptr(unsafe.Pointer(ptr("Path Home: Structural Reentry Assistance | Start Here"))), WS_OVERLAPPEDWINDOW, 90, 65, 1130, 850, 0, 0, inst, 0)
	if h == 0 {
		log.Printf("startup: main window creation failed: %v", createErr)
		alert("Path Home could not create its main window.")
		return 1
	}
	hwnd = h
	font = proc(pCreateFont, 21, 0, 0, 0, 480, 0, 0, 0, 1, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	buttonFont = proc(pCreateFont, 18, 0, 0, 0, 600, 0, 0, 0, 1, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	largeFont = proc(pCreateFont, 28, 0, 0, 0, 650, 0, 0, 0, 1, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	titleFont = proc(pCreateFont, 33, 0, 0, 0, 730, 0, 0, 0, 1, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	defer func() {
		releaseBitmaps()
		for _, obj := range []uintptr{font, buttonFont, largeFont, titleFont, brush} {
			if obj != 0 { proc(pDeleteObject, obj) }
		}
	}()
	vaultLastActivity=time.Now()
	if proc(user32.NewProc("SetTimer"),h,1,1000,0)==0 {log.Print("vault idle timer unavailable")}
	defer proc(user32.NewProc("KillTimer"),h,1)
	if proc(wtsRegister,h,0)!=0 {defer proc(wtsUnregister,h)} else {log.Print("session lock notification unavailable")}
	log.Print("startup: rendering home")
	draw()
	proc(pShowWindow, h, 5)
	proc(pUpdateWindow, h)
	log.Print("startup: entering GUI message loop")
	var m message
	for {
		v, _, err := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(v) == -1 {
			log.Printf("message loop failure: %v", err)
			return 1
		}
		if int32(v) == 0 { return 0 }
		if (m.Message>=0x0100 && m.Message<=0x0109)||(m.Message>=0x0200 && m.Message<=0x020E){vaultLastActivity=time.Now()}
		proc(pTranslateMessage, uintptr(unsafe.Pointer(&m)))
		proc(pDispatchMessage, uintptr(unsafe.Pointer(&m)))
	}
}
