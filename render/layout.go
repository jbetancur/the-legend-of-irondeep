package render

const (
	ScreenW = 1920
	ScreenH = 1080

	// The viewport is rendered internally at ViewportW x ViewportH (see
	// viewport.go) and then scaled to this on-screen display size.
	viewportDisplayW = 1280
	viewportDisplayH = 720

	viewportScreenX = (ScreenW - viewportDisplayW) / 2 // 320
	viewportScreenY = 0

	// Party bar: full width along the bottom, in the remaining space.
	partyBarY = viewportDisplayH           // 720
	partyBarH = ScreenH - viewportDisplayH // 360
)
