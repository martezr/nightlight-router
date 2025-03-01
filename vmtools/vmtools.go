package vmtools

import (
	"log"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
)

func QueryGuestInfo(vsphereConfig *rpcvmx.Config, data string) (output string) {
	if out, err := vsphereConfig.String(data, ""); err != nil {
		log.Fatalf("ERROR: string failed with %s", err)
	} else {
		log.Printf("guest info: %s - %s\n", data, out)
		return out
	}
	return "unable to parse value"
}
