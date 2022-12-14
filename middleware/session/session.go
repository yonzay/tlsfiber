package session

import (
	"sync"
	"time"

	"github.com/yonzay/tlsfiber"
	"github.com/yonzay/tlsfiber/internal/gotiny"
	"github.com/yonzay/tlsfiber/utils"
	"github.com/yonzay/fasthttp"
)

type Session struct {
	id     string     // session id
	fresh  bool       // if new session
	ctx    *fiber.Ctx // fiber context
	config *Store     // store configuration
	data   *data      // key value data
}

var sessionPool = sync.Pool{
	New: func() interface{} {
		return new(Session)
	},
}

func acquireSession() *Session {
	s := sessionPool.Get().(*Session)
	if s.data == nil {
		s.data = acquireData()
	}
	s.fresh = true
	return s
}

func releaseSession(s *Session) {
	s.id = ""
	s.ctx = nil
	s.config = nil
	if s.data != nil {
		s.data.Reset()
	}
	sessionPool.Put(s)
}

// Fresh is true if the current session is new
func (s *Session) Fresh() bool {
	return s.fresh
}

// ID returns the session id
func (s *Session) ID() string {
	return s.id
}

// Get will return the value
func (s *Session) Get(key string) interface{} {
	// Better safe than sorry
	if s.data == nil {
		return nil
	}
	return s.data.Get(key)
}

// Set will update or create a new key value
func (s *Session) Set(key string, val interface{}) {
	// Better safe than sorry
	if s.data == nil {
		return
	}
	s.data.Set(key, val)
}

// Delete will delete the value
func (s *Session) Delete(key string) {
	// Better safe than sorry
	if s.data == nil {
		return
	}
	s.data.Delete(key)
}

// Destroy will delete the session from Storage and expire session cookie
func (s *Session) Destroy() error {
	// Better safe than sorry
	if s.data == nil {
		return nil
	}

	// Reset local data
	s.data.Reset()

	// Use external Storage if exist
	if err := s.config.Storage.Delete(s.id); err != nil {
		return err
	}

	// Expire cookie
	s.delCookie()
	return nil
}

// Regenerate generates a new session id and delete the old one from Storage
func (s *Session) Regenerate() error {

	// Delete old id from storage
	if err := s.config.Storage.Delete(s.id); err != nil {
		return err
	}

	// Create new ID
	s.id = s.config.KeyGenerator()

	return nil
}

// Save will update the storage and client cookie
func (s *Session) Save() error {

	// Better safe than sorry
	if s.data == nil {
		return nil
	}

	// Create cookie with the session ID if fresh
	if s.fresh {
		s.setCookie()
	}

	// Don't save to Storage if no data is available
	if s.data.Len() <= 0 {
		return nil
	}

	// Convert data to bytes
	mux.Lock()
	data := gotiny.Marshal(&s.data)
	mux.Unlock()

	// pass raw bytes with session id to provider
	if err := s.config.Storage.Set(s.id, data, s.config.Expiration); err != nil {
		return err
	}

	// Release session
	// TODO: It's not safe to use the Session after called Save()
	releaseSession(s)

	return nil
}

func (s *Session) setCookie() {
	fcookie := fasthttp.AcquireCookie()
	fcookie.SetKey(s.config.CookieName)
	fcookie.SetValue(s.id)
	fcookie.SetPath(s.config.CookiePath)
	fcookie.SetDomain(s.config.CookieDomain)
	fcookie.SetMaxAge(int(s.config.Expiration.Seconds()))
	fcookie.SetExpire(time.Now().Add(s.config.Expiration))
	fcookie.SetSecure(s.config.CookieSecure)
	fcookie.SetHTTPOnly(s.config.CookieHTTPOnly)

	switch utils.ToLower(s.config.CookieSameSite) {
	case "strict":
		fcookie.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	case "none":
		fcookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	default:
		fcookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	}

	s.ctx.Response().Header.SetCookie(fcookie)
	fasthttp.ReleaseCookie(fcookie)
}

func (s *Session) delCookie() {
	s.ctx.Request().Header.DelCookie(s.config.CookieName)
	s.ctx.Response().Header.DelCookie(s.config.CookieName)

	fcookie := fasthttp.AcquireCookie()
	fcookie.SetKey(s.config.CookieName)
	fcookie.SetPath(s.config.CookiePath)
	fcookie.SetDomain(s.config.CookieDomain)
	fcookie.SetMaxAge(-1)
	fcookie.SetExpire(time.Now().Add(-1 * time.Minute))
	fcookie.SetSecure(s.config.CookieSecure)
	fcookie.SetHTTPOnly(s.config.CookieHTTPOnly)

	switch utils.ToLower(s.config.CookieSameSite) {
	case "strict":
		fcookie.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	case "none":
		fcookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	default:
		fcookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	}

	s.ctx.Response().Header.SetCookie(fcookie)
	fasthttp.ReleaseCookie(fcookie)
}
