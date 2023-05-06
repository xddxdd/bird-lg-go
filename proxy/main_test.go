package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestHasAccessNotConfigured(t *testing.T) {
	setting.allowedIPs = []net.IP{}
	assert.Equal(t, hasAccess("whatever"), true)
}

func TestHasAccessAllowIPv4(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("1.2.3.4")}
	assert.Equal(t, hasAccess("1.2.3.4:4321"), true)
}

func TestHasAccessDenyIPv4(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("4.3.2.1")}
	assert.Equal(t, hasAccess("1.2.3.4:4321"), false)
}

func TestHasAccessAllowIPv6(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("2001:db8::1")}
	assert.Equal(t, hasAccess("[2001:db8::1]:4321"), true)
}

func TestHasAccessAllowIPv6DifferentForm(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("2001:0db8::1")}
	assert.Equal(t, hasAccess("[2001:db8::1]:4321"), true)
}

func TestHasAccessDenyIPv6(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("2001:db8::2")}
	assert.Equal(t, hasAccess("[2001:db8::1]:4321"), false)
}

func TestHasAccessBadClientIP(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("1.2.3.4")}
	assert.Equal(t, hasAccess("not an IP"), false)
}

func TestHasAccessBadClientIPPort(t *testing.T) {
	setting.allowedIPs = []net.IP{net.ParseIP("1.2.3.4")}
	assert.Equal(t, hasAccess("not an IP:not a port"), false)
}

func TestAccessHandlerAllow(t *testing.T) {
	baseHandler := http.NotFoundHandler()
	wrappedHandler := accessHandler(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/mock", nil)
	r.RemoteAddr = "1.2.3.4:4321"
	w := httptest.NewRecorder()

	setting.allowedIPs = []net.IP{net.ParseIP("1.2.3.4")}

	wrappedHandler.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestAccessHandlerDeny(t *testing.T) {
	baseHandler := http.NotFoundHandler()
	wrappedHandler := accessHandler(baseHandler)

	r := httptest.NewRequest(http.MethodGet, "/mock", nil)
	r.RemoteAddr = "1.2.3.4:4321"
	w := httptest.NewRecorder()

	setting.allowedIPs = []net.IP{net.ParseIP("4.3.2.1")}

	wrappedHandler.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
}
