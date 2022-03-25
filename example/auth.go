package example

import (
	"encoding/json"
	"fmt"
	"github.com/dlshle/aghs/server"
	"github.com/dlshle/aghs/utils"
	"sync"
	"time"
)

const AUTH_HEADER = "Authorization"
const BEARER_TOKEN = "Bearer "
const STUDENT_ID_CONTEXT_KEY = "student_id"

type authInfo struct {
	id        string
	expiresAt time.Time
}

type authMiddleware struct {
	tokenMap map[string]authInfo //token -- authInfo
	lock     *sync.RWMutex
}

func NewAuthMiddleware() authMiddleware {
	return authMiddleware{
		make(map[string]authInfo),
		new(sync.RWMutex),
	}
}

func (m authMiddleware) withWrite(cb func()) {
	m.lock.Lock()
	defer m.lock.Unlock()
	cb()
}

func (m authMiddleware) authorize(r server.Request) (server.Request, server.ServiceError) {
	authHeader := r.Header().Get(AUTH_HEADER)
	if len(authHeader) <= len(BEARER_TOKEN) {
		return r, server.ForbiddenError("invalid token")
	}
	token := authHeader[len(BEARER_TOKEN):]
	studentId, err := m.validate(token)
	if err != nil {
		return r, server.ForbiddenError(err.Error())
	}
	r.Context().RegisterContext(STUDENT_ID_CONTEXT_KEY, studentId)
	return r, nil
}

func (m authMiddleware) validate(token string) (id string, err error) {
	var info authInfo
	info, exists := m.getIdByToken(token)
	if !exists {
		info, err = m.decodeToken(token)
		if err != nil {
			return
		}
	}
	if info.expiresAt.After(time.Now()) {
		err = fmt.Errorf("token expired")
		return
	}
	id = info.id
	return
}

func (m authMiddleware) grantToken(student Student) string {
	authInfo := authInfo{student.Id, time.Now().Add(time.Minute * 30)}
	marshalled, _ := json.Marshal(authInfo)
	m.withWrite(func() {
		m.tokenMap[student.Id] = authInfo
	})
	return string(marshalled)
}

func (m authMiddleware) decodeToken(token string) (info authInfo, err error) {
	decoded, err := utils.DecodeBase64(token)
	if err != nil {
		return
	}
	err = json.Unmarshal(decoded, &info)
	return
}

func (m authMiddleware) getIdByToken(token string) (info authInfo, exists bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	info, exists = m.tokenMap[token]
	return
}
