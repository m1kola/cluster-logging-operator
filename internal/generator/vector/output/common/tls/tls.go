package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type TLSConf struct {
	Component          string
	ID                 string
	NeedsEnabled       bool
	InsecureSkipVerify bool
	TlsMinVersion      string
	CipherSuites       string
	CAFilePath         string
	CertPath           string
	KeyPath            string
	PassPhrase         string
}

func New(id string, spec *obs.OutputTLSSpec, secrets helpers.Secrets, op framework.Options) TLSConf {
	conf := TLSConf{
		Component: "sinks",
		ID:        id,
	}
	if spec != nil {
		conf.CAFilePath = ConfigMapOrSecretPath(spec.CA)
		conf.CertPath = ConfigMapOrSecretPath(spec.Certificate)
		conf.KeyPath = SecretPath(spec.Key)
		conf.PassPhrase = secrets.AsString(spec.KeyPassphrase)
		conf.InsecureSkipVerify = spec.InsecureSkipVerify
	}
	setTLSProfileFromOptions(&conf, op)
	if conf.CipherSuites != "" || conf.TlsMinVersion != "" || spec != nil {
		conf.NeedsEnabled = true
	}
	return conf
}

func ConfigMapOrSecretPath(resource *obs.ConfigMapOrSecretKey) string {
	if resource == nil {
		return ""
	}
	if resource.Secret != nil {
		return common.SecretPath(resource.Secret.Name, resource.Key)
	} else if resource.ConfigMap != nil {
		return common.ConfigmapPath(resource.ConfigMap.Name, resource.Key)
	}
	return ""
}

func SecretPath(resource *obs.SecretKey) string {
	if resource == nil || resource.Secret == nil {
		return ""
	}
	return common.SecretPath(resource.Secret.Name, resource.Key)
}

func setTLSProfileFromOptions(t *TLSConf, op framework.Options) {
	if version, found := op[framework.MinTLSVersion]; found {
		t.TlsMinVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		t.CipherSuites = ciphers.(string)
	}
}

func (t TLSConf) Name() string {
	return "vectorTLS"
}

func (t TLSConf) Template() string {
	if !t.NeedsEnabled {
		return `{{define "vectorTLS" -}}{{end}}`
	}
	return `
{{define "vectorTLS" -}}
[{{.Component}}.{{.ID}}.tls]
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}
{{- if .InsecureSkipVerify }}
verify_certificate = false
verify_hostname = false
{{- end }}
{{- if and .KeyPath .CertPath }}
key_file = {{ .KeyPath }}
crt_file = {{ .CertPath }}
{{- end }}
{{- if .CAFilePath }}
ca_file = {{ .CAFilePath }}
{{- end }}
{{- if .PassPhrase }}
key_pass = "{{ .PassPhrase }}"
{{- end }}
{{ end }}`
}
