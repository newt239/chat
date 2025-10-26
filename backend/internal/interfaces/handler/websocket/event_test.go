package websocket

import (
	"encoding/json"
	"testing"
)

func TestParseClientMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EventType
		wantErr bool
	}{
		{
			name:    "join_channel event",
			input:   `{"type":"join_channel","payload":{"channel_id":"ch1"}}`,
			want:    EventTypeJoinChannel,
			wantErr: false,
		},
		{
			name:    "leave_channel event",
			input:   `{"type":"leave_channel","payload":{"channel_id":"ch1"}}`,
			want:    EventTypeLeaveChannel,
			wantErr: false,
		},
		{
			name:    "post_message event",
			input:   `{"type":"post_message","payload":{"channel_id":"ch1","body":"hello"}}`,
			want:    EventTypePostMessage,
			wantErr: false,
		},
		{
			name:    "typing event",
			input:   `{"type":"typing","payload":{"channel_id":"ch1"}}`,
			want:    EventTypeTyping,
			wantErr: false,
		},
		{
			name:    "update_read_state event",
			input:   `{"type":"update_read_state","payload":{"channel_id":"ch1","message_id":"msg1"}}`,
			want:    EventTypeUpdateReadState,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseClientMessage([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClientMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && msg.Type != tt.want {
				t.Errorf("ParseClientMessage() type = %v, want %v", msg.Type, tt.want)
			}
		})
	}
}

func TestSendServerMessage(t *testing.T) {
	tests := []struct {
		name    string
		evtType EventType
		payload interface{}
		wantErr bool
	}{
		{
			name:    "new_message event",
			evtType: EventTypeNewMessage,
			payload: NewMessagePayload{
				ChannelID: "ch1",
				Message:   map[string]interface{}{"id": "msg1", "body": "hello"},
			},
			wantErr: false,
		},
		{
			name:    "unread_count event",
			evtType: EventTypeUnreadCount,
			payload: UnreadCountPayload{
				ChannelID:   "ch1",
				UnreadCount: 5,
			},
			wantErr: false,
		},
		{
			name:    "ack event",
			evtType: EventTypeAck,
			payload: AckPayload{
				Type:    EventTypeJoinChannel,
				Success: true,
				Message: "joined successfully",
			},
			wantErr: false,
		},
		{
			name:    "error event",
			evtType: EventTypeError,
			payload: ErrorPayload{
				Code:    "INTERNAL_ERROR",
				Message: "something went wrong",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SendServerMessage(tt.evtType, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendServerMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// JSONとしてパース可能か確認
				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("SendServerMessage() produced invalid JSON: %v", err)
				}
				// typeフィールドが正しいか確認
				if result["type"] != string(tt.evtType) {
					t.Errorf("SendServerMessage() type = %v, want %v", result["type"], tt.evtType)
				}
			}
		})
	}
}
