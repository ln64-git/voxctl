package audio

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

type MPRISController struct {
	player *AudioPlayer
	conn   *dbus.Conn
}

func NewMPRISController(player *AudioPlayer) *MPRISController {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Printf("Failed to connect to session bus: %v\n", err)
		return nil
	}

	mprisController := &MPRISController{
		player: player,
		conn:   conn,
	}

	mprisController.registerMPRIS()

	return mprisController
}

func (mc *MPRISController) registerMPRIS() {
	reply, err := mc.conn.RequestName("org.mpris.MediaPlayer2.audio", dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Printf("Failed to request name on session bus: %v\n", err)
		return
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Println("Name already taken")
		return
	}

	node := &introspect.Node{
		Name: "/org/mpris/MediaPlayer2",
		Interfaces: []introspect.Interface{
			{
				Name: "org.mpris.MediaPlayer2",
				Methods: []introspect.Method{
					{Name: "Quit"},
				},
				Properties: []introspect.Property{
					{Name: "CanQuit", Type: "b", Access: "read"},
					{Name: "CanRaise", Type: "b", Access: "read"},
					{Name: "HasTrackList", Type: "b", Access: "read"},
					{Name: "Identity", Type: "s", Access: "read"},
					{Name: "SupportedUriSchemes", Type: "as", Access: "read"},
					{Name: "SupportedMimeTypes", Type: "as", Access: "read"},
				},
			},
			{
				Name: "org.mpris.MediaPlayer2.Player",
				Methods: []introspect.Method{
					{Name: "Play"},
					{Name: "Pause"},
					{Name: "Stop"},
					{Name: "Next"},
					{Name: "Previous"},
				},
				Properties: []introspect.Property{
					{Name: "PlaybackStatus", Type: "s", Access: "read"},
					{Name: "LoopStatus", Type: "s", Access: "readwrite"},
					{Name: "Rate", Type: "d", Access: "readwrite"},
					{Name: "Shuffle", Type: "b", Access: "readwrite"},
					{Name: "Metadata", Type: "a{sv}", Access: "read"},
					{Name: "Volume", Type: "d", Access: "readwrite"},
					{Name: "Position", Type: "x", Access: "read"},
					{Name: "MinimumRate", Type: "d", Access: "read"},
					{Name: "MaximumRate", Type: "d", Access: "read"},
					{Name: "CanGoNext", Type: "b", Access: "read"},
					{Name: "CanGoPrevious", Type: "b", Access: "read"},
					{Name: "CanPlay", Type: "b", Access: "read"},
					{Name: "CanPause", Type: "b", Access: "read"},
					{Name: "CanSeek", Type: "b", Access: "read"},
					{Name: "CanControl", Type: "b", Access: "read"},
				},
			},
		},
	}

	mc.conn.Export(mc, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	mc.conn.Export(mc, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")
	mc.conn.Export(introspect.NewIntrospectable(node), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")
}

func (mc *MPRISController) Play() {
	mc.player.Play()
}

func (mc *MPRISController) Pause() {
	mc.player.Pause()
}

func (mc *MPRISController) Stop() {
	mc.player.Stop()
}

func (mc *MPRISController) Next() {
	// Handle skipping to next track
}

func (mc *MPRISController) Previous() {
	// Handle skipping to previous track
}

func (mc *MPRISController) Quit() {
	// Handle quitting the application
}

func (mc *MPRISController) CanQuit() bool          { return true }
func (mc *MPRISController) CanRaise() bool         { return false }
func (mc *MPRISController) HasTrackList() bool     { return false }
func (mc *MPRISController) Identity() string       { return "Audio Player" }
func (mc *MPRISController) SupportedUriSchemes() []string { return []string{} }
func (mc *MPRISController) SupportedMimeTypes() []string  { return []string{} }
func (mc *MPRISController) PlaybackStatus() string { return "Playing" }
func (mc *MPRISController) LoopStatus() string     { return "None" }
func (mc *MPRISController) SetLoopStatus(status string)   {}
func (mc *MPRISController) Rate() float64         { return 1.0 }
func (mc *MPRISController) SetRate(rate float64)  {}
func (mc *MPRISController) Shuffle() bool         { return false }
func (mc *MPRISController) SetShuffle(shuffle bool) {}
func (mc *MPRISController) Metadata() map[string]dbus.Variant { return map[string]dbus.Variant{} }
func (mc *MPRISController) Volume() float64        { return 1.0 }
func (mc *MPRISController) SetVolume(volume float64) {}
func (mc *MPRISController) Position() int64        { return 0 }
func (mc *MPRISController) MinimumRate() float64   { return 1.0 }
func (mc *MPRISController) MaximumRate() float64   { return 1.0 }
func (mc *MPRISController) CanGoNext() bool        { return false }
func (mc *MPRISController) CanGoPrevious() bool    { return false }
func (mc *MPRISController) CanPlay() bool          { return true }
func (mc *MPRISController) CanPause() bool         { return true }
func (mc *MPRISController) CanSeek() bool          { return false }
func (mc *MPRISController) CanControl() bool       { return true }