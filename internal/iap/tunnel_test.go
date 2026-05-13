package iap

import (
	"encoding/binary"
	"testing"
)

func TestFrameEncoding(t *testing.T) {
	// Test DATA frame encoding.
	payload := []byte("hello world")
	frame := make([]byte, headerLen+len(payload))
	binary.BigEndian.PutUint16(frame[0:2], tagData)
	binary.BigEndian.PutUint32(frame[2:6], uint32(len(payload)))
	copy(frame[headerLen:], payload)

	// Verify tag.
	tag := binary.BigEndian.Uint16(frame[0:2])
	if tag != tagData {
		t.Errorf("tag = 0x%04x, want 0x%04x", tag, tagData)
	}

	// Verify length.
	dataLen := binary.BigEndian.Uint32(frame[2:6])
	if dataLen != uint32(len(payload)) {
		t.Errorf("data length = %d, want %d", dataLen, len(payload))
	}

	// Verify payload.
	decoded := frame[headerLen : headerLen+int(dataLen)]
	if string(decoded) != "hello world" {
		t.Errorf("payload = %q, want %q", decoded, "hello world")
	}
}

func TestACKFrameEncoding(t *testing.T) {
	var bytesRecv uint64 = 123456789

	frame := make([]byte, tagLen+8)
	binary.BigEndian.PutUint16(frame[0:2], tagACK)
	binary.BigEndian.PutUint64(frame[2:10], bytesRecv)

	tag := binary.BigEndian.Uint16(frame[0:2])
	if tag != tagACK {
		t.Errorf("tag = 0x%04x, want 0x%04x", tag, tagACK)
	}

	ack := binary.BigEndian.Uint64(frame[2:10])
	if ack != bytesRecv {
		t.Errorf("ack = %d, want %d", ack, bytesRecv)
	}
}

func TestConnectSuccessFrameParsing(t *testing.T) {
	sid := []byte("session-id-abc-123")
	frame := make([]byte, headerLen+len(sid))
	binary.BigEndian.PutUint16(frame[0:2], tagConnectSuccessSID)
	binary.BigEndian.PutUint32(frame[2:6], uint32(len(sid)))
	copy(frame[headerLen:], sid)

	tag := binary.BigEndian.Uint16(frame[0:2])
	if tag != tagConnectSuccessSID {
		t.Errorf("tag = 0x%04x, want 0x%04x", tag, tagConnectSuccessSID)
	}

	dataLen := binary.BigEndian.Uint32(frame[2:6])
	parsedSID := frame[headerLen : headerLen+int(dataLen)]
	if string(parsedSID) != "session-id-abc-123" {
		t.Errorf("sid = %q, want %q", parsedSID, "session-id-abc-123")
	}
}

func TestTagConstants(t *testing.T) {
	if tagConnectSuccessSID != 0x0001 {
		t.Error("tagConnectSuccessSID wrong")
	}
	if tagReconnectSuccessACK != 0x0002 {
		t.Error("tagReconnectSuccessACK wrong")
	}
	if tagData != 0x0004 {
		t.Error("tagData wrong")
	}
	if tagACK != 0x0007 {
		t.Error("tagACK wrong")
	}
	if maxDataFrameSize != 16384 {
		t.Error("maxDataFrameSize wrong")
	}
}
