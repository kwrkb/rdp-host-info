// Package msgid defines the message identifiers used across diag, hostinfo,
// and render. It has no dependencies, so diag/hostinfo can reference IDs
// without importing the msg catalog that resolves them to language text.
package msgid

type ID string

const (
	InternalError     ID = "internal_error"
	InternalErrorHint ID = "internal_error_hint"

	EditionUnknown         ID = "edition_unknown"
	EditionUnknownHint     ID = "edition_unknown_hint"
	EditionUnsupported     ID = "edition_unsupported"
	EditionUnsupportedHint ID = "edition_unsupported_hint"
	EditionSupported       ID = "edition_supported"

	RDPEnabledUnknown     ID = "rdp_enabled_unknown"
	RDPEnabledUnknownHint ID = "rdp_enabled_unknown_hint"
	RDPDisabled           ID = "rdp_disabled"
	RDPDisabledHint       ID = "rdp_disabled_hint"
	RDPEnabled            ID = "rdp_enabled"

	ServiceUnknown        ID = "service_unknown"
	ServiceUnknownHint    ID = "service_unknown_hint"
	ServiceNotRunning     ID = "service_not_running"
	ServiceNotRunningHint ID = "service_not_running_hint"
	ServiceRunning        ID = "service_running"

	FirewallUnknown           ID = "firewall_unknown"
	FirewallUnknownHint       ID = "firewall_unknown_hint"
	FirewallOK                ID = "firewall_ok"
	FirewallOKFallback        ID = "firewall_ok_fallback"
	FirewallPublicBlocked     ID = "firewall_public_blocked"
	FirewallPublicBlockedHint ID = "firewall_public_blocked_hint"
	FirewallBlocked           ID = "firewall_blocked"
	FirewallBlockedHint       ID = "firewall_blocked_hint"

	PortUnknown             ID = "port_unknown"
	PortUnknownHint         ID = "port_unknown_hint"
	PortNotListening        ID = "port_not_listening"
	PortNotListeningAssumed ID = "port_not_listening_assumed"
	PortNotListeningHint    ID = "port_not_listening_hint"
	PortListening           ID = "port_listening"
	PortListeningAssumed    ID = "port_listening_assumed"

	GroupUnknown       ID = "group_unknown"
	GroupUnknownHint   ID = "group_unknown_hint"
	GroupNotMember     ID = "group_not_member"
	GroupNotMemberHint ID = "group_not_member_hint"
	GroupOKAdmin       ID = "group_ok_admin"
	GroupOKRDU         ID = "group_ok_rdu"
	GroupOKBoth        ID = "group_ok_both"
	GroupOKHint        ID = "group_ok_hint"

	SleepUnknown     ID = "sleep_unknown"
	SleepUnknownHint ID = "sleep_unknown_hint"
	SleepNever       ID = "sleep_never"
	SleepWarnBoth    ID = "sleep_warn_both"
	SleepWarnAC      ID = "sleep_warn_ac"
	SleepWarnDC      ID = "sleep_warn_dc"
	SleepWarnHint    ID = "sleep_warn_hint"

	NoteAzureADUPN        ID = "note_azuread_upn"
	NoteMSAPassword       ID = "note_msa_password"
	NoteMaybeMSA          ID = "note_maybe_msa"
	LabelLocalAccount     ID = "label_local_account"
	LabelDomainAccount    ID = "label_domain_account"
	LabelAzureADAccount   ID = "label_azuread_account"
	LabelMicrosoftAccount ID = "label_microsoft_account"

	HeaderHostInfo      ID = "header_host_info"
	HeaderPCName        ID = "header_pc_name"
	HeaderWindows       ID = "header_windows"
	HeaderConnAddr      ID = "header_conn_addr"
	LabelLocalIP        ID = "label_local_ip"
	LabelTailscaleIP    ID = "label_tailscale_ip"
	HeaderUsername      ID = "header_username"
	HeaderRecommended   ID = "header_recommended"
	HeaderStatus        ID = "header_status"
	SuffixAdminRequired ID = "suffix_admin_required"
	Unknown             ID = "unknown"

	UsageText ID = "usage_text"
)

// All lists every ID that must have a catalog entry in every supported
// language. Used by internal/msg's completeness test.
var All = []ID{
	InternalError, InternalErrorHint,
	EditionUnknown, EditionUnknownHint, EditionUnsupported, EditionUnsupportedHint, EditionSupported,
	RDPEnabledUnknown, RDPEnabledUnknownHint, RDPDisabled, RDPDisabledHint, RDPEnabled,
	ServiceUnknown, ServiceUnknownHint, ServiceNotRunning, ServiceNotRunningHint, ServiceRunning,
	FirewallUnknown, FirewallUnknownHint, FirewallOK, FirewallOKFallback,
	FirewallPublicBlocked, FirewallPublicBlockedHint, FirewallBlocked, FirewallBlockedHint,
	PortUnknown, PortUnknownHint, PortNotListening, PortNotListeningAssumed, PortNotListeningHint,
	PortListening, PortListeningAssumed,
	GroupUnknown, GroupUnknownHint, GroupNotMember, GroupNotMemberHint,
	GroupOKAdmin, GroupOKRDU, GroupOKBoth, GroupOKHint,
	SleepUnknown, SleepUnknownHint, SleepNever, SleepWarnBoth, SleepWarnAC, SleepWarnDC, SleepWarnHint,
	NoteAzureADUPN, NoteMSAPassword, NoteMaybeMSA,
	LabelLocalAccount, LabelDomainAccount, LabelAzureADAccount, LabelMicrosoftAccount,
	HeaderHostInfo, HeaderPCName, HeaderWindows, HeaderConnAddr, LabelLocalIP, LabelTailscaleIP,
	HeaderUsername, HeaderRecommended, HeaderStatus, SuffixAdminRequired, Unknown,
	UsageText,
}
