package render

import (
	"image"
	_ "image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Assets struct {
	FrontWall *ebiten.Image
	SideWallL *ebiten.Image
	SideWallR *ebiten.Image
	Door      *ebiten.Image
	Floor     *ebiten.Image
	Ceiling   *ebiten.Image
	Monsters  map[string]*ebiten.Image
}

func loadTexture(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("load texture %s: %v", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("decode texture %s: %v", path, err)
	}
	return ebiten.NewImageFromImage(img)
}

func NewAssets(wallset string) *Assets {
	dir := "assets/textures/" + wallset + "/"
	sideWall := loadTexture(dir + "wall_side.png")
	return &Assets{
		FrontWall: loadTexture(dir + "wall_front.png"),
		SideWallL: sideWall,
		SideWallR: sideWall,
		Door:      loadTexture(dir + "door.png"),
		Floor:     loadTexture(dir + "floor.png"),
		Ceiling:   loadTexture(dir + "ceiling.png"),
		Monsters:  loadMonsterSprites(),
	}
}

func loadMonsterSprites() map[string]*ebiten.Image {
	sprites := make(map[string]*ebiten.Image)
	entries, err := os.ReadDir("assets/monsters")
	if err != nil {
		return sprites
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) > 4 && name[len(name)-4:] == ".png" {
			key := name[:len(name)-4]
			sprites[key] = loadTexture("assets/monsters/" + name)
		}
	}
	return sprites
}
