// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package appengine

import (
	"strings"
	"time"

	pb "appengine_internal/app_identity"
	"appengine_internal"
	modpb "appengine_internal/modules"
)

// AppID returns the application ID for the current application.
// The string will be a plain application ID (e.g. "appid"), with a
// domain prefix for custom domain deployments (e.g. "example.com:appid").
func AppID(c Context) string { return appengine_internal.AppID(c.FullyQualifiedAppID()) }

// BackendInstance returns the name and index of the current backend instance,
// or "", -1 if this is not a backend instance.
func BackendInstance(c Context) (name string, index int) {
	index = appengine_internal.BackendInstance()
	if index == -1 {
		return
	}
	name = VersionID(c)
	if i := strings.Index(name, "."); i > -1 {
		name = name[:i]
	}
	return
}

// BackendHostname returns the standard hostname of the specified backend.
// If index is -1, BackendHostname returns the load-balancing hostname for
// the backend.
func BackendHostname(c Context, name string, index int) string {
	return appengine_internal.BackendHostname(c, name, index)
}

// DefaultVersionHostname returns the standard hostname of the default version
// of the current application (e.g. "my-app.appspot.com"). This is suitable for
// use in constructing URLs.
func DefaultVersionHostname(c Context) string {
	return appengine_internal.DefaultVersionHostname(c.Request())
}

// ModuleName returns the module name of the current instance.
func ModuleName(c Context) string {
	return appengine_internal.ModuleName(c.Request())
}

// ModuleHostname returns a hostname of a module instance.
// If module is the empty string, it refers to the module of the current instance.
// If version is empty, it refers to the version of the current instance if valid,
// or the default version of the module of the current instance.
// If instance is empty, ModuleHostname returns the load-balancing hostname.
func ModuleHostname(c Context, module, version, instance string) (string, error) {
	req := &modpb.GetHostnameRequest{}
	if module != "" {
		req.Module = &module
	}
	if version != "" {
		req.Version = &version
	}
	if instance != "" {
		req.Instance = &instance
	}
	res := &modpb.GetHostnameResponse{}
	if err := c.Call("modules", "GetHostname", req, res, nil); err != nil {
		return "", err
	}
	return *res.Hostname, nil
}

// VersionID returns the version ID for the current application.
// It will be of the form "X.Y", where X is specified in app.yaml,
// and Y is a number generated when each version of the app is uploaded.
// It does not include a module name.
func VersionID(c Context) string { return appengine_internal.VersionID(c.Request()) }

// InstanceID returns a mostly-unique identifier for this instance.
func InstanceID() string { return appengine_internal.InstanceID() }

// Datacenter returns an identifier for the datacenter that the instance is running in.
func Datacenter() string { return appengine_internal.Datacenter() }

// ServerSoftware returns the App Engine release version.
// In production, it looks like "Google App Engine/X.Y.Z".
// In the development appserver, it looks like "Development/X.Y".
func ServerSoftware() string { return appengine_internal.ServerSoftware() }

// RequestID returns a string that uniquely identifies the request.
func RequestID(c Context) string { return appengine_internal.RequestID(c.Request()) }

// AccessToken generates an OAuth2 access token for the specified scopes on
// behalf of service account of this application. This token will expire after
// the returned time.
func AccessToken(c Context, scopes ...string) (token string, expiry time.Time, err error) {
	req := &pb.GetAccessTokenRequest{Scope: scopes}
	res := &pb.GetAccessTokenResponse{}

	err = c.Call("app_identity_service", "GetAccessToken", req, res, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	return res.GetAccessToken(), time.Unix(res.GetExpirationTime(), 0), nil
}

// Certificate represents a public certificate for the app.
type Certificate struct {
	KeyName string
	Data    []byte // PEM-encoded X.509 certificate
}

// PublicCertificates retrieves the public certificates for the app.
// They can be used to verify a signature returned by SignBytes.
func PublicCertificates(c Context) ([]Certificate, error) {
	req := &pb.GetPublicCertificateForAppRequest{}
	res := &pb.GetPublicCertificateForAppResponse{}
	if err := c.Call("app_identity_service", "GetPublicCertificatesForApp", req, res, nil); err != nil {
		return nil, err
	}
	var cs []Certificate
	for _, pc := range res.PublicCertificateList {
		cs = append(cs, Certificate{
			KeyName: pc.GetKeyName(),
			Data:    []byte(pc.GetX509CertificatePem()),
		})
	}
	return cs, nil
}

// ServiceAccount returns a string representing the service account name, in
// the form of an email address (typically app_id@appspot.gserviceaccount.com).
func ServiceAccount(c Context) (string, error) {
	req := &pb.GetServiceAccountNameRequest{}
	res := &pb.GetServiceAccountNameResponse{}

	err := c.Call("app_identity_service", "GetServiceAccountName", req, res, nil)
	if err != nil {
		return "", err
	}
	return res.GetServiceAccountName(), err
}

// SignBytes signs bytes using a private key unique to your application.
func SignBytes(c Context, bytes []byte) (keyName string, signature []byte, err error) {
	req := &pb.SignForAppRequest{BytesToSign: bytes}
	res := &pb.SignForAppResponse{}

	if err := c.Call("app_identity_service", "SignForApp", req, res, nil); err != nil {
		return "", nil, err
	}
	return res.GetKeyName(), res.GetSignatureBytes(), nil
}

func init() {
	appengine_internal.RegisterErrorCodeMap("app_identity_service", pb.AppIdentityServiceError_ErrorCode_name)
	appengine_internal.RegisterErrorCodeMap("modules", modpb.ModulesServiceError_ErrorCode_name)
}
