package tts

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	edgeURL = "wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=6A5AA1D4EAFF4E9FB37E23D68491D6F4"
)

func randomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getTimestamp() string {
	return time.Now().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

type EdgeTTS struct {
	Voice string
	Rate  string // e.g. "+0%", "+10%"
	Pitch string // e.g. "+0Hz"
}

func NewEdgeTTS() *EdgeTTS {
	return &EdgeTTS{
		Voice: "zh-CN-XiaoxiaoNeural",
		Rate:  "+0%",
		Pitch: "+0Hz",
	}
}

func (e *EdgeTTS) Synthesize(text string) ([]byte, error) {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(edgeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to edge tts: %v", err)
	}
	defer conn.Close()

	// 1. Send Config
	reqID := randomID()
	configMsg := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":\"false\",\"wordBoundaryEnabled\":\"false\"},\"outputFormat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}", getTimestamp())
	if err := conn.WriteMessage(websocket.TextMessage, []byte(configMsg)); err != nil {
		return nil, err
	}

	// 2. Send SSML
	ssml := fmt.Sprintf("<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='zh-CN'><voice name='%s'><prosody rate='%s' pitch='%s'>%s</prosody></voice></speak>", e.Voice, e.Rate, e.Pitch, text)
	ssmlMsg := fmt.Sprintf("X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%s\r\nPath:ssml\r\n\r\n%s", reqID, getTimestamp(), ssml)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlMsg)); err != nil {
		return nil, err
	}

	// 3. Receive Audio
	var audioData []byte
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType == websocket.TextMessage {
			text := string(msg)
			if strings.Contains(text, "Path:turn.end") {
				break
			}
		} else if msgType == websocket.BinaryMessage {
			// Binary message format: header (text) + binary
			// Find the position of "Path:audio\r\n"
			headerEnd := strings.Index(string(msg), "\r\n\r\n")
			if headerEnd != -1 {
				// Check if it is audio
				header := string(msg[:headerEnd])
				if strings.Contains(header, "Path:audio") {
					audioData = append(audioData, msg[headerEnd+4:]...)
				}
			}
		}
	}

	if len(audioData) == 0 {
		return nil, fmt.Errorf("no audio data received")
	}

	return audioData, nil
}
