package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errWriter is an io.Writer that always returns an error.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

// limitedReader wraps a reader to provide only n bytes before EOF.
func limitedBuf(data []byte, n int) io.Reader {
	if n > len(data) {
		n = len(data)
	}
	return bytes.NewReader(data[:n])
}

// encodePayload encodes a packet to bytes (the payload only, not framed).
func encodePayload(t *testing.T, pkt Packet) []byte {
	t.Helper()
	var buf bytes.Buffer
	err := pkt.Encode(&buf)
	assert.NoError(t, err)
	return buf.Bytes()
}

// helper: roundtrip a packet through WritePacket/ReadPacket and return the decoded result.
func roundtrip(t *testing.T, reg *PacketRegistry, pkt Packet) Packet {
	t.Helper()
	var buf bytes.Buffer
	err := WritePacket(&buf, pkt)
	assert.NoError(t, err)
	got, err := ReadPacket(&buf, reg)
	assert.NoError(t, err)
	return got
}

// ---------- Login packets ----------

func TestLoginRequest_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &LoginRequest{Username: "Steve"}
	got := roundtrip(t, reg, orig).(*LoginRequest)
	assert.Equal(t, orig.Username, got.Username)
	assert.Equal(t, byte(0x01), got.ID())
}

func TestLoginResponse_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &LoginResponse{EntityID: 42, SpawnX: 1.5, SpawnY: 64.0, SpawnZ: -3.25}
	got := roundtrip(t, reg, orig).(*LoginResponse)
	assert.Equal(t, orig.EntityID, got.EntityID)
	assert.Equal(t, orig.SpawnX, got.SpawnX)
	assert.Equal(t, orig.SpawnY, got.SpawnY)
	assert.Equal(t, orig.SpawnZ, got.SpawnZ)
	assert.Equal(t, byte(0x02), got.ID())
}

func TestDisconnect_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &Disconnect{Reason: "Server shutting down"}
	got := roundtrip(t, reg, orig).(*Disconnect)
	assert.Equal(t, orig.Reason, got.Reason)
	assert.Equal(t, byte(0x03), got.ID())
}

// ---------- Play packets ----------

func TestPlayerPosition_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &PlayerPosition{X: 10.5, Y: 64.0, Z: -20.75, Yaw: 90.0, Pitch: -45.0, OnGround: true}
	got := roundtrip(t, reg, orig).(*PlayerPosition)
	assert.Equal(t, orig.X, got.X)
	assert.Equal(t, orig.Y, got.Y)
	assert.Equal(t, orig.Z, got.Z)
	assert.Equal(t, orig.Yaw, got.Yaw)
	assert.Equal(t, orig.Pitch, got.Pitch)
	assert.Equal(t, orig.OnGround, got.OnGround)
	assert.Equal(t, byte(0x10), got.ID())
}

func TestPlayerAction_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &PlayerAction{ActionType: 1, BlockX: 100, BlockY: 64, BlockZ: -200, Face: 3}
	got := roundtrip(t, reg, orig).(*PlayerAction)
	assert.Equal(t, orig.ActionType, got.ActionType)
	assert.Equal(t, orig.BlockX, got.BlockX)
	assert.Equal(t, orig.BlockY, got.BlockY)
	assert.Equal(t, orig.BlockZ, got.BlockZ)
	assert.Equal(t, orig.Face, got.Face)
	assert.Equal(t, byte(0x11), got.ID())
}

func TestChatMessage_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &ChatMessage{Sender: "Alex", Message: "Hello, world!"}
	got := roundtrip(t, reg, orig).(*ChatMessage)
	assert.Equal(t, orig.Sender, got.Sender)
	assert.Equal(t, orig.Message, got.Message)
	assert.Equal(t, byte(0x12), got.ID())
}

// ---------- World packets ----------

func TestChunkData_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	orig := &ChunkData{ChunkX: 5, ChunkZ: -3, Data: data}
	got := roundtrip(t, reg, orig).(*ChunkData)
	assert.Equal(t, orig.ChunkX, got.ChunkX)
	assert.Equal(t, orig.ChunkZ, got.ChunkZ)
	assert.Equal(t, orig.Data, got.Data)
	assert.Equal(t, byte(0x20), got.ID())
}

func TestBlockChange_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &BlockChange{X: 10, Y: 64, Z: -20, BlockID: 4}
	got := roundtrip(t, reg, orig).(*BlockChange)
	assert.Equal(t, orig.X, got.X)
	assert.Equal(t, orig.Y, got.Y)
	assert.Equal(t, orig.Z, got.Z)
	assert.Equal(t, orig.BlockID, got.BlockID)
	assert.Equal(t, byte(0x21), got.ID())
}

func TestSpawnEntity_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &SpawnEntity{EntityID: 99, EntityType: 2, X: 1.0, Y: 2.0, Z: 3.0, Yaw: 45.0, Pitch: -10.0}
	got := roundtrip(t, reg, orig).(*SpawnEntity)
	assert.Equal(t, orig.EntityID, got.EntityID)
	assert.Equal(t, orig.EntityType, got.EntityType)
	assert.Equal(t, orig.X, got.X)
	assert.Equal(t, orig.Y, got.Y)
	assert.Equal(t, orig.Z, got.Z)
	assert.Equal(t, orig.Yaw, got.Yaw)
	assert.Equal(t, orig.Pitch, got.Pitch)
	assert.Equal(t, byte(0x22), got.ID())
}

// ---------- System packets ----------

func TestKeepAlive_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &KeepAlive{Timestamp: 1234567890}
	got := roundtrip(t, reg, orig).(*KeepAlive)
	assert.Equal(t, orig.Timestamp, got.Timestamp)
	assert.Equal(t, byte(0x30), got.ID())
}

func TestPing_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &Ping{Timestamp: 9876543210}
	got := roundtrip(t, reg, orig).(*Ping)
	assert.Equal(t, orig.Timestamp, got.Timestamp)
	assert.Equal(t, byte(0x31), got.ID())
}

func TestPong_Roundtrip(t *testing.T) {
	reg := DefaultRegistry()
	orig := &Pong{Timestamp: 1111111111}
	got := roundtrip(t, reg, orig).(*Pong)
	assert.Equal(t, orig.Timestamp, got.Timestamp)
	assert.Equal(t, byte(0x32), got.ID())
}

// ---------- Frame format verification ----------

func TestFrameFormat(t *testing.T) {
	pkt := &LoginRequest{Username: "Test"}
	var buf bytes.Buffer
	err := WritePacket(&buf, pkt)
	assert.NoError(t, err)

	raw := buf.Bytes()

	// First byte is packet ID
	assert.Equal(t, byte(0x01), raw[0], "first byte should be packet ID")

	// Next 4 bytes are payload length (big-endian uint32)
	length := binary.BigEndian.Uint32(raw[1:5])

	// "Test" is 4 bytes, plus 2-byte uint16 length prefix = 6 bytes payload
	assert.Equal(t, uint32(6), length, "payload length should be 6")

	// Remaining bytes are the payload
	assert.Equal(t, int(length), len(raw)-5, "remaining bytes should match payload length")

	// Total frame: 1 (ID) + 4 (length) + 6 (payload) = 11 bytes
	assert.Equal(t, 11, len(raw), "total frame size should be 11")
}

// ---------- Unknown packet ID error ----------

func TestUnknownPacketID(t *testing.T) {
	reg := DefaultRegistry()

	// Manually construct a frame with an unregistered ID (0xFF)
	var buf bytes.Buffer
	buf.WriteByte(0xFF)
	_ = binary.Write(&buf, binary.BigEndian, uint32(0)) // zero-length payload

	_, err := ReadPacket(&buf, reg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown packet ID: 0xFF")
}

// ---------- Registry Create error ----------

func TestRegistryCreateUnknown(t *testing.T) {
	reg := NewPacketRegistry()
	_, err := reg.Create(0xAB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown packet ID: 0xAB")
}

// ---------- Empty ChunkData roundtrip ----------

func TestChunkData_EmptyData(t *testing.T) {
	reg := DefaultRegistry()
	orig := &ChunkData{ChunkX: 0, ChunkZ: 0, Data: []byte{}}
	got := roundtrip(t, reg, orig).(*ChunkData)
	assert.Equal(t, orig.ChunkX, got.ChunkX)
	assert.Equal(t, orig.ChunkZ, got.ChunkZ)
	assert.Equal(t, orig.Data, got.Data)
}

// ---------- Bool false roundtrip ----------

func TestPlayerPosition_OnGroundFalse(t *testing.T) {
	reg := DefaultRegistry()
	orig := &PlayerPosition{X: 0, Y: 0, Z: 0, Yaw: 0, Pitch: 0, OnGround: false}
	got := roundtrip(t, reg, orig).(*PlayerPosition)
	assert.False(t, got.OnGround)
}

// ---------- String with special characters ----------

func TestChatMessage_Unicode(t *testing.T) {
	reg := DefaultRegistry()
	orig := &ChatMessage{Sender: "Player1", Message: "Hello! Emoji test: unicode chars and such"}
	got := roundtrip(t, reg, orig).(*ChatMessage)
	assert.Equal(t, orig.Sender, got.Sender)
	assert.Equal(t, orig.Message, got.Message)
}

// ---------- DefaultRegistry completeness ----------

func TestDefaultRegistryAllPackets(t *testing.T) {
	reg := DefaultRegistry()

	ids := []byte{0x01, 0x02, 0x03, 0x10, 0x11, 0x12, 0x20, 0x21, 0x22, 0x30, 0x31, 0x32}
	for _, id := range ids {
		pkt, err := reg.Create(id)
		assert.NoError(t, err, "should create packet for ID 0x%02X", id)
		assert.Equal(t, id, pkt.ID(), "created packet should have ID 0x%02X", id)
	}
}

// ---------- io helpers coverage ----------

func TestIOHelpers_Float64(t *testing.T) {
	var buf bytes.Buffer
	err := WriteFloat64(&buf, 3.141592653589793)
	assert.NoError(t, err)
	val, err := ReadFloat64(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 3.141592653589793, val)
}

func TestIOHelpers_Int64(t *testing.T) {
	var buf bytes.Buffer
	err := WriteInt64(&buf, -9223372036854775808)
	assert.NoError(t, err)
	val, err := ReadInt64(&buf)
	assert.NoError(t, err)
	assert.Equal(t, int64(-9223372036854775808), val)
}

func TestIOHelpers_Uint16(t *testing.T) {
	var buf bytes.Buffer
	err := WriteUint16(&buf, 65535)
	assert.NoError(t, err)
	val, err := ReadUint16(&buf)
	assert.NoError(t, err)
	assert.Equal(t, uint16(65535), val)
}

func TestIOHelpers_Byte(t *testing.T) {
	var buf bytes.Buffer
	err := WriteByte_(&buf, 0xAB)
	assert.NoError(t, err)
	val, err := ReadByte_(&buf)
	assert.NoError(t, err)
	assert.Equal(t, byte(0xAB), val)
}

func TestIOHelpers_Bool(t *testing.T) {
	var buf bytes.Buffer
	err := WriteBool(&buf, true)
	assert.NoError(t, err)
	val, err := ReadBool(&buf)
	assert.NoError(t, err)
	assert.True(t, val)

	err = WriteBool(&buf, false)
	assert.NoError(t, err)
	val, err = ReadBool(&buf)
	assert.NoError(t, err)
	assert.False(t, val)
}

func TestIOHelpers_String(t *testing.T) {
	var buf bytes.Buffer
	err := WriteString(&buf, "hello")
	assert.NoError(t, err)
	val, err := ReadString(&buf)
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestIOHelpers_EmptyString(t *testing.T) {
	var buf bytes.Buffer
	err := WriteString(&buf, "")
	assert.NoError(t, err)
	val, err := ReadString(&buf)
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestIOHelpers_Int32(t *testing.T) {
	var buf bytes.Buffer
	err := WriteInt32(&buf, -2147483648)
	assert.NoError(t, err)
	val, err := ReadInt32(&buf)
	assert.NoError(t, err)
	assert.Equal(t, int32(-2147483648), val)
}

func TestIOHelpers_Uint32(t *testing.T) {
	var buf bytes.Buffer
	err := WriteUint32(&buf, 4294967295)
	assert.NoError(t, err)
	val, err := ReadUint32(&buf)
	assert.NoError(t, err)
	assert.Equal(t, uint32(4294967295), val)
}

func TestIOHelpers_Float32(t *testing.T) {
	var buf bytes.Buffer
	err := WriteFloat32(&buf, 1.5)
	assert.NoError(t, err)
	val, err := ReadFloat32(&buf)
	assert.NoError(t, err)
	assert.Equal(t, float32(1.5), val)
}

// ---------- Truncated decode error paths ----------
// These tests verify that Decode returns errors when the reader runs out of data
// at each field boundary, covering the intermediate error branches.

func TestLoginResponse_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &LoginResponse{EntityID: 1, SpawnX: 2, SpawnY: 3, SpawnZ: 4})
	// Truncate at various points: 0 (EntityID), 4 (SpawnX), 8 (SpawnY), 12 (SpawnZ)
	cuts := []int{0, 3, 5, 9}
	for _, n := range cuts {
		p := &LoginResponse{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestPlayerPosition_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &PlayerPosition{X: 1, Y: 2, Z: 3, Yaw: 4, Pitch: 5, OnGround: true})
	// Fields: X(4), Y(4), Z(4), Yaw(4), Pitch(4), OnGround(1) = 21 bytes
	cuts := []int{0, 3, 5, 9, 13, 17}
	for _, n := range cuts {
		p := &PlayerPosition{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestPlayerAction_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &PlayerAction{ActionType: 1, BlockX: 2, BlockY: 3, BlockZ: 4, Face: 5})
	// Fields: ActionType(1), BlockX(4), BlockY(4), BlockZ(4), Face(1) = 14 bytes
	cuts := []int{0, 2, 6, 10}
	for _, n := range cuts {
		p := &PlayerAction{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestChatMessage_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &ChatMessage{Sender: "A", Message: "B"})
	// Fields: Sender(2+1=3), Message(2+1=3) = 6 bytes
	cuts := []int{0, 2}
	for _, n := range cuts {
		p := &ChatMessage{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestChunkData_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &ChunkData{ChunkX: 1, ChunkZ: 2, Data: []byte{0xAA, 0xBB}})
	// Fields: ChunkX(4), ChunkZ(4), DataLen(4), Data(2) = 14 bytes
	cuts := []int{0, 3, 5, 9}
	for _, n := range cuts {
		p := &ChunkData{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestBlockChange_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &BlockChange{X: 1, Y: 2, Z: 3, BlockID: 4})
	// Fields: X(4), Y(4), Z(4), BlockID(2) = 14 bytes
	cuts := []int{0, 3, 5, 9}
	for _, n := range cuts {
		p := &BlockChange{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

func TestSpawnEntity_DecodeErrors(t *testing.T) {
	full := encodePayload(t, &SpawnEntity{EntityID: 1, EntityType: 2, X: 3, Y: 4, Z: 5, Yaw: 6, Pitch: 7})
	// Fields: EntityID(4), EntityType(1), X(4), Y(4), Z(4), Yaw(4), Pitch(4) = 25 bytes
	cuts := []int{0, 3, 5, 6, 10, 14, 18}
	for _, n := range cuts {
		p := &SpawnEntity{}
		err := p.Decode(limitedBuf(full, n))
		assert.Error(t, err, "should fail with %d bytes", n)
	}
}

// ---------- Encode error paths (using errWriter) ----------

func TestLoginResponse_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &LoginResponse{EntityID: 1, SpawnX: 2, SpawnY: 3, SpawnZ: 4}
	assert.Error(t, p.Encode(w))
}

func TestPlayerPosition_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &PlayerPosition{X: 1, Y: 2, Z: 3, Yaw: 4, Pitch: 5, OnGround: true}
	assert.Error(t, p.Encode(w))
}

func TestPlayerAction_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &PlayerAction{ActionType: 1, BlockX: 2, BlockY: 3, BlockZ: 4, Face: 5}
	assert.Error(t, p.Encode(w))
}

func TestChatMessage_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &ChatMessage{Sender: "A", Message: "B"}
	assert.Error(t, p.Encode(w))
}

func TestChunkData_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &ChunkData{ChunkX: 1, ChunkZ: 2, Data: []byte{0xAA}}
	assert.Error(t, p.Encode(w))
}

func TestBlockChange_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &BlockChange{X: 1, Y: 2, Z: 3, BlockID: 4}
	assert.Error(t, p.Encode(w))
}

func TestSpawnEntity_EncodeErrors(t *testing.T) {
	w := errWriter{}
	p := &SpawnEntity{EntityID: 1, EntityType: 2, X: 3, Y: 4, Z: 5, Yaw: 6, Pitch: 7}
	assert.Error(t, p.Encode(w))
}

// ---------- WritePacket / ReadPacket error paths ----------

func TestWritePacket_EncodeError(t *testing.T) {
	// WritePacket with errWriter -- error writing ID
	w := errWriter{}
	pkt := &LoginRequest{Username: "test"}
	err := WritePacket(w, pkt)
	assert.Error(t, err)
}

func TestReadPacket_TruncatedFrame(t *testing.T) {
	reg := DefaultRegistry()

	// Empty reader
	_, err := ReadPacket(bytes.NewReader([]byte{}), reg)
	assert.Error(t, err)

	// Only ID byte, no length
	_, err = ReadPacket(bytes.NewReader([]byte{0x01}), reg)
	assert.Error(t, err)

	// ID + length but missing payload
	var buf bytes.Buffer
	buf.WriteByte(0x01)
	_ = binary.Write(&buf, binary.BigEndian, uint32(10))
	_, err = ReadPacket(&buf, reg)
	assert.Error(t, err)
}

func TestReadPacket_DecodeError(t *testing.T) {
	reg := DefaultRegistry()

	// Valid frame for LoginRequest but payload is truncated (1 byte instead of needed)
	var buf bytes.Buffer
	buf.WriteByte(0x01)
	_ = binary.Write(&buf, binary.BigEndian, uint32(1))
	buf.WriteByte(0x00) // 1 byte payload, but LoginRequest needs at least 2 for string length
	_, err := ReadPacket(&buf, reg)
	assert.Error(t, err)
}

// ---------- IO helpers read error paths ----------

func TestReadString_Error(t *testing.T) {
	// Only 1 byte when 2 are needed for length prefix
	_, err := ReadString(bytes.NewReader([]byte{0x00}))
	assert.Error(t, err)
}

func TestReadString_TruncatedData(t *testing.T) {
	// Length says 5 but only 2 bytes of data
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint16(5))
	buf.Write([]byte("ab"))
	_, err := ReadString(&buf)
	assert.Error(t, err)
}

func TestWriteString_LengthError(t *testing.T) {
	err := WriteString(errWriter{}, "test")
	assert.Error(t, err)
}

func TestWriteString_DataError(t *testing.T) {
	// A writer that succeeds for the first 2 bytes (length) then fails
	w := &limitWriter{limit: 2}
	err := WriteString(w, "test")
	assert.Error(t, err)
}

// limitWriter allows writing up to `limit` bytes then fails.
type limitWriter struct {
	written int
	limit   int
}

func (w *limitWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.written
	if remaining <= 0 {
		return 0, errors.New("write limit reached")
	}
	if len(p) <= remaining {
		w.written += len(p)
		return len(p), nil
	}
	w.written += remaining
	return remaining, errors.New("write limit reached")
}

func TestReadFloat32_Error(t *testing.T) {
	_, err := ReadFloat32(bytes.NewReader([]byte{0x00}))
	assert.Error(t, err)
}

func TestReadFloat64_Error(t *testing.T) {
	_, err := ReadFloat64(bytes.NewReader([]byte{0x00}))
	assert.Error(t, err)
}

func TestReadInt32_Error(t *testing.T) {
	_, err := ReadInt32(bytes.NewReader([]byte{0x00}))
	assert.Error(t, err)
}

func TestReadUint16_Error(t *testing.T) {
	_, err := ReadUint16(bytes.NewReader([]byte{}))
	assert.Error(t, err)
}

func TestReadUint32_Error(t *testing.T) {
	_, err := ReadUint32(bytes.NewReader([]byte{0x00}))
	assert.Error(t, err)
}

func TestReadByte_Error(t *testing.T) {
	_, err := ReadByte_(bytes.NewReader([]byte{}))
	assert.Error(t, err)
}

func TestReadBool_Error(t *testing.T) {
	_, err := ReadBool(bytes.NewReader([]byte{}))
	assert.Error(t, err)
}

func TestReadInt64_Error(t *testing.T) {
	_, err := ReadInt64(bytes.NewReader([]byte{0x00, 0x01}))
	assert.Error(t, err)
}

// ---------- Encode with partially-failing writers ----------
// These use limitWriter to hit intermediate encode error branches.

func TestLoginResponse_EncodePartialErrors(t *testing.T) {
	p := &LoginResponse{EntityID: 1, SpawnX: 2, SpawnY: 3, SpawnZ: 4}
	// Fail after EntityID (4 bytes), during SpawnX
	w := &limitWriter{limit: 4}
	assert.Error(t, p.Encode(w))

	// Fail after EntityID + SpawnX (8 bytes), during SpawnY
	w = &limitWriter{limit: 8}
	assert.Error(t, p.Encode(w))

	// Fail after EntityID + SpawnX + SpawnY (12 bytes), during SpawnZ
	w = &limitWriter{limit: 12}
	assert.Error(t, p.Encode(w))
}

func TestPlayerPosition_EncodePartialErrors(t *testing.T) {
	p := &PlayerPosition{X: 1, Y: 2, Z: 3, Yaw: 4, Pitch: 5, OnGround: true}
	// Each float32 is 4 bytes, bool is 1 byte
	limits := []int{4, 8, 12, 16, 20}
	for _, limit := range limits {
		w := &limitWriter{limit: limit}
		assert.Error(t, p.Encode(w), "should fail with limit %d", limit)
	}
}

func TestPlayerAction_EncodePartialErrors(t *testing.T) {
	p := &PlayerAction{ActionType: 1, BlockX: 2, BlockY: 3, BlockZ: 4, Face: 5}
	// ActionType(1), BlockX(4), BlockY(4), BlockZ(4), Face(1)
	limits := []int{1, 5, 9, 13}
	for _, limit := range limits {
		w := &limitWriter{limit: limit}
		assert.Error(t, p.Encode(w), "should fail with limit %d", limit)
	}
}

func TestChatMessage_EncodePartialErrors(t *testing.T) {
	p := &ChatMessage{Sender: "A", Message: "B"}
	// Sender: 2 (len) + 1 (data) = 3 bytes, then Message fails
	w := &limitWriter{limit: 3}
	assert.Error(t, p.Encode(w))
}

func TestChunkData_EncodePartialErrors(t *testing.T) {
	p := &ChunkData{ChunkX: 1, ChunkZ: 2, Data: []byte{0xAA, 0xBB}}
	// ChunkX(4), ChunkZ(4), DataLen(4), Data(2)
	limits := []int{4, 8, 12}
	for _, limit := range limits {
		w := &limitWriter{limit: limit}
		assert.Error(t, p.Encode(w), "should fail with limit %d", limit)
	}
}

func TestBlockChange_EncodePartialErrors(t *testing.T) {
	p := &BlockChange{X: 1, Y: 2, Z: 3, BlockID: 4}
	// X(4), Y(4), Z(4), BlockID(2)
	limits := []int{4, 8, 12}
	for _, limit := range limits {
		w := &limitWriter{limit: limit}
		assert.Error(t, p.Encode(w), "should fail with limit %d", limit)
	}
}

func TestSpawnEntity_EncodePartialErrors(t *testing.T) {
	p := &SpawnEntity{EntityID: 1, EntityType: 2, X: 3, Y: 4, Z: 5, Yaw: 6, Pitch: 7}
	// EntityID(4), EntityType(1), X(4), Y(4), Z(4), Yaw(4), Pitch(4)
	limits := []int{4, 5, 9, 13, 17, 21}
	for _, limit := range limits {
		w := &limitWriter{limit: limit}
		assert.Error(t, p.Encode(w), "should fail with limit %d", limit)
	}
}
