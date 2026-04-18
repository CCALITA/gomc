package ui

import (
	"fmt"
	"time"

	"github.com/fanxiyao/gomc/internal/input"
)

const (
	chatMaxMessages = 50
	chatLineHeight  = 16.0
	chatPadX        = 8.0
	chatPadY        = 8.0
	chatInputHeight = 24.0
	chatBgAlpha     = 0.6
)

// Key code for backspace editing.
const keyBackspace = 259

// ChatMessage represents a single message in the chat log.
type chatMessage struct {
	Sender string
	Text   string
	Time   time.Time
}

// String returns a formatted "[Sender] Text" representation, or just
// the Text when there is no sender (system messages).
func (m chatMessage) String() string {
	if m.Sender == "" {
		return m.Text
	}
	return fmt.Sprintf("[%s] %s", m.Sender, m.Text)
}

// onSendFunc is called when the player submits a line in the chat.
// The line may be a plain message or a slash command.
type onSendFunc func(text string)

// ChatScreen implements the Screen interface for an in-game chat
// overlay. It collects keyboard input into a text buffer, displays
// recent messages, and invokes OnSend when the player presses Enter.
type ChatScreen struct {
	Messages []chatMessage
	InputBuf []rune

	onSend  onSendFunc
	onClose func()
	closed  bool
}

// NewChatScreen creates a ChatScreen ready to receive input.
// onSend is called with the input text when Enter is pressed.
// onClose is called when the chat is dismissed (Escape or after send).
func NewChatScreen(onSend onSendFunc, onClose func()) *ChatScreen {
	return &ChatScreen{
		Messages: make([]chatMessage, 0, chatMaxMessages),
		InputBuf: make([]rune, 0, 128),
		onSend:   onSend,
		onClose:  onClose,
	}
}

// AddMessage appends a message to the chat log, evicting the oldest
// message when the log exceeds chatMaxMessages.
func (c *ChatScreen) AddMessage(sender, text string) {
	msg := chatMessage{
		Sender: sender,
		Text:   text,
		Time:   time.Now(),
	}
	if len(c.Messages) >= chatMaxMessages {
		// Drop oldest by creating a new slice (immutable style).
		trimmed := make([]chatMessage, len(c.Messages)-1, chatMaxMessages)
		copy(trimmed, c.Messages[1:])
		c.Messages = append(trimmed, msg)
	} else {
		c.Messages = append(c.Messages, msg)
	}
}

// Update handles per-frame input: Escape to close, Enter to send,
// Backspace to delete, and printable key presses to append characters.
func (c *ChatScreen) Update(inp *input.Manager, _ float64) {
	if c.closed {
		return
	}

	if inp.IsKeyJustPressed(input.KeyEscape) {
		c.Close()
		return
	}

	if inp.IsKeyJustPressed(input.KeyEnter) {
		c.send()
		return
	}

	if inp.IsKeyJustPressed(keyBackspace) {
		if len(c.InputBuf) > 0 {
			c.InputBuf = c.InputBuf[:len(c.InputBuf)-1]
		}
	}
}

// HandleKey processes a raw key code. For the chat screen, printable
// ASCII characters (32-126) are appended to the input buffer.
func (c *ChatScreen) HandleKey(key int) {
	if c.closed {
		return
	}
	if key >= 32 && key <= 126 {
		c.InputBuf = append(c.InputBuf, rune(key))
	}
}

// HandleChar processes a Unicode character from GLFW's char callback.
// This is the primary method for receiving typed text.
func (c *ChatScreen) HandleChar(ch rune) {
	if c.closed {
		return
	}
	if ch >= 32 && ch <= 126 {
		c.InputBuf = append(c.InputBuf, ch)
	}
}

// Draw renders the chat overlay: a dark semi-transparent background,
// recent messages scrolling upward, and the input box at the bottom.
func (c *ChatScreen) Draw(r *UIRenderer) {
	panelWidth := r.ScreenWidth * 0.5
	if panelWidth < 300 {
		panelWidth = 300
	}
	panelHeight := r.ScreenHeight * 0.4
	panelX := float32(0)
	panelY := r.ScreenHeight - panelHeight

	// Dark semi-transparent background.
	r.DrawRect(panelX, panelY, panelWidth, panelHeight, 0, 0, 0, chatBgAlpha)

	// Input box at the bottom of the panel.
	inputY := r.ScreenHeight - chatInputHeight - chatPadY
	r.DrawRect(panelX+chatPadX, inputY, panelWidth-2*chatPadX, chatInputHeight, 0.15, 0.15, 0.15, 0.8)

	// Input text with cursor.
	inputText := string(c.InputBuf) + "_"
	r.DrawText(panelX+chatPadX+4, inputY+6, inputText, 1.0, 1, 1, 1)

	// Messages above the input box, drawn bottom-to-top.
	maxVisible := int((inputY - panelY - chatPadY) / chatLineHeight)
	startIdx := 0
	if len(c.Messages) > maxVisible {
		startIdx = len(c.Messages) - maxVisible
	}

	visibleMessages := c.Messages[startIdx:]
	for i, msg := range visibleMessages {
		lineY := inputY - float32(len(visibleMessages)-i)*chatLineHeight
		if lineY < panelY {
			continue
		}
		r.DrawText(panelX+chatPadX, lineY, msg.String(), 1.0, 0.9, 0.9, 0.9)
	}
}

// IsOverlay returns true: the chat is drawn on top of the game world.
func (c *ChatScreen) IsOverlay() bool {
	return true
}

// IsClosed reports whether the chat screen has been dismissed.
func (c *ChatScreen) IsClosed() bool {
	return c.closed
}

// InputText returns the current contents of the input buffer as a string.
func (c *ChatScreen) InputText() string {
	return string(c.InputBuf)
}

// Close dismisses the chat screen.
func (c *ChatScreen) Close() {
	if c.closed {
		return
	}
	c.closed = true
	if c.onClose != nil {
		c.onClose()
	}
}

// send submits the current input buffer and closes the chat.
func (c *ChatScreen) send() {
	text := string(c.InputBuf)
	c.InputBuf = c.InputBuf[:0]

	if text != "" && c.onSend != nil {
		c.onSend(text)
	}
	c.Close()
}
