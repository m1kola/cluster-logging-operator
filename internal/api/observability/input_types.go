package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/utils/set"
)

func MaxRecordsPerSecond(input obs.InputSpec) (int64, bool) {
	if input.Application != nil &&
		input.Application.Tuning != nil &&
		input.Application.Tuning.RateLimitPerContainer != nil {
		return Threshold(input.Application.Tuning.RateLimitPerContainer)
	}
	return 0, false
}

func Threshold(ls *obs.LimitSpec) (int64, bool) {
	if ls == nil {
		return 0, false
	}
	return ls.MaxRecordsPerSecond, true
}

type Inputs []obs.InputSpec

// ConfigmapNames returns a unique set of unordered configmap names
func (inputs Inputs) ConfigmapNames() []string {
	names := set.New[string]()
	for _, i := range inputs {
		if i.Receiver != nil && i.Receiver.TLS != nil {
			names.Insert(ConfigmapsForTLS(obs.TLSSpec(*i.Receiver.TLS))...)
		}
	}
	return names.UnsortedList()
}

// SecretNames returns a unique set of unordered secret names
func (inputs Inputs) SecretNames() []string {
	secrets := set.New[string]()
	for _, i := range inputs {
		if i.Receiver != nil && i.Receiver.TLS != nil {
			secrets.Insert(SecretsForTLS(obs.TLSSpec(*i.Receiver.TLS))...)
		}
	}
	return secrets.UnsortedList()
}

func (inputs Inputs) HasJournalSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceNode)) {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasContainerSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeApplication {
			return true
		}
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceContainer)) {
			return true
		}
	}
	return false
}
func (inputs Inputs) HasAnyAuditSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeAudit && i.Audit != nil {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasAuditSource(logSource obs.AuditSource) bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeAudit && i.Audit != nil && (set.New(i.Audit.Sources...).Has(logSource) || len(i.Audit.Sources) == 0) {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasReceiverSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeReceiver && i.Receiver != nil {
			return true
		}
	}
	return false
}
