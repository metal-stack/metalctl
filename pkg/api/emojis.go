package api

const (
	Ambulance   = "🚑"
	Exclamation = "❗"
	Bark        = "🚧"
	Loop        = "⭕"
	Lock        = "🔒"
	Question    = "❓"
	Skull       = "💀"
	VPN         = "🛡️ "
)

func EmojiHelpText() string {
	return `
Meaning of the emojis:

🚧 Machine is reserved. Reserved machines are not considered for random allocation until the reservation flag is removed.
🔒 Machine is locked. Locked machines can not be deleted until the lock is removed.
💀 Machine is dead. The metal-api does not receive any events from this machine.
❗ Machine has a last event error. The machine has recently encountered an error during the provisioning lifecycle.
❓ Machine is in unknown condition. The metal-api does not receive phoned home events anymore or has never booted successfully.
⭕ Machine is in a provisioning crash loop. Flag can be reset through an API-triggered reboot or when the machine reaches the phoned home state.
🚑 Machine reclaim has failed. The machine was deleted but it is not going back into the available machine pool.
🛡️  Machine is connected to our VPN, ssh access only possible via this VPN.
`
}
