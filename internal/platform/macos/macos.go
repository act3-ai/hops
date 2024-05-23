//revive:disable:exported
package macos

type Version struct {
	OSName        string
	Name          string
	Version       string
	DarwinVersion int
}

type Darwin int

const (
	DarwinUnknown Darwin = iota - 1
	DarwinJaguar  Darwin = iota + 5
	DarwinPanther
	DarwinTiger
	DarwinLeopard
	DarwinSnowLeopard
	DarwinLion
	DarwinMountainLion
	DarwinMavericks
	DarwinYosemite
	DarwinElCapitan
	DarwinSierra
	DarwinHighSierra
	DarwinMojave
	DarwinCatalina
	DarwinBigSur
	DarwinMonterey
	DarwinVentura
	DarwinSonoma
)

var Sonoma = Version{
	OSName:        "macOS",
	Name:          "Sonoma",
	Version:       "14",
	DarwinVersion: 23,
}

var Ventura = Version{
	OSName:        "macOS",
	Name:          "Ventura",
	Version:       "13",
	DarwinVersion: 22,
}

var Monterey = Version{
	OSName:        "macOS",
	Name:          "Monterey",
	Version:       "12",
	DarwinVersion: 21,
}

var BigSur = Version{
	OSName:        "macOS",
	Name:          "Big Sur",
	Version:       "11",
	DarwinVersion: 20,
}

var Catalina = Version{
	OSName:        "macOS",
	Name:          "Catalina",
	Version:       "10.15",
	DarwinVersion: 19,
}

var Mojave = Version{
	OSName:        "macOS",
	Name:          "Mojave",
	Version:       "10.14",
	DarwinVersion: 18,
}

var HighSierra = Version{
	OSName:        "macOS",
	Name:          "High Sierra",
	Version:       "10.13",
	DarwinVersion: 17,
}

var Sierra = Version{
	OSName:        "macOS",
	Name:          "Sierra",
	Version:       "10.12",
	DarwinVersion: 16,
}

var ElCapitan = Version{
	OSName:        "OS X",
	Name:          "El Capitan",
	Version:       "10.11",
	DarwinVersion: 15,
}

// SupportsARM reports whether the version supports ARM applications.
func (v Version) SupportsARM() bool {
	return v.DarwinVersion >= 20
}

// Supports64Bit reports whether the version supports 64-bit applications.
func (v Version) Supports64Bit() bool {
	return v.DarwinVersion >= 7
}

// Supports32Bit reports whether the version supports 32-bit applications.
func (v Version) Supports32Bit() bool {
	return v.DarwinVersion < 19
}
