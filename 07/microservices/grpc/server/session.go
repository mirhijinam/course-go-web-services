package main

import (
	"context"
	"fmt"
	"math/rand"
	"microservices/grpc/session"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sessKeyLen = 10

type SessionManager struct {
	session.UnimplementedAuthCheckerServer
	mu       sync.RWMutex
	sessions map[string]*session.Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		UnimplementedAuthCheckerServer: session.UnimplementedAuthCheckerServer{},
		mu:                             sync.RWMutex{},
		sessions:                       map[string]*session.Session{},
	}
}

func (sm *SessionManager) Create(ctx context.Context, in *session.Session) (*session.SessionID, error) {
	fmt.Println("call Create", in)
	id := &session.SessionID{ID: RandStringRunes(sessKeyLen)}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[id.ID] = in
	return id, nil
}

func (sm *SessionManager) Check(ctx context.Context, in *session.SessionID) (*session.Session, error) {
	fmt.Println("call Check", in)
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sess, ok := sm.sessions[in.ID]; ok {
		return sess, nil
	}
	return nil, status.Errorf(codes.NotFound, "session not found")
}

func (sm *SessionManager) Delete(ctx context.Context, in *session.SessionID) (*session.Nothing, error) {
	fmt.Println("call Delete", in)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, in.ID)
	return &session.Nothing{Dummy: true}, nil
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
