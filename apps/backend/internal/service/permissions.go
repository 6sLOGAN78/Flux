package service

// Capability describes a closed server-side workspace permission.
type Capability string

// Workspace capabilities are denied by default for unknown roles or operations.
const (
	roleOwner                  = "owner"
	roleAdmin                  = "admin"
	CapabilityRead  Capability = "read"
	CapabilityWrite Capability = "write"
	CapabilityTeam  Capability = "team"
)

const (
	roleMember = "member"
	roleViewer = "viewer"
)

// Allows grants permissions only to the documented Flux roles.
func Allows(role string, capability Capability) bool {
	switch role {
	case roleOwner, roleAdmin:
		return capability == CapabilityRead || capability == CapabilityWrite || capability == CapabilityTeam
	case roleMember:
		return capability == CapabilityRead || capability == CapabilityWrite
	case roleViewer:
		return capability == CapabilityRead
	default:
		return false
	}
}

// AllowsRoleChange checks BOTH current and proposed target roles. Owner-count
// invariants additionally require exclusive workspace locking at mutation time.
func AllowsRoleChange(actorRole, currentRole, proposedRole string) bool {
	if !Allows(currentRole, CapabilityRead) || !Allows(proposedRole, CapabilityRead) {
		return false
	}
	if actorRole == roleOwner {
		return true
	}
	return actorRole == roleAdmin && (currentRole == roleMember || currentRole == roleViewer) &&
		(proposedRole == roleMember || proposedRole == roleViewer)
}
