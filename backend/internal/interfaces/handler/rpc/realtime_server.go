package rpc

import "github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"

type RealtimeServer struct {
	chatv1connect.UnimplementedRealtimeServiceHandler
}
