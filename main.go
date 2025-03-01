package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	sysctl "github.com/lorenzosaino/go-sysctl"
	"github.com/martezr/nightlight-router/vmtools"
	"github.com/smarty/cproxy/v2"
	"github.com/vishvananda/netlink"
	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

// guestinfo.nightlight-router.ip
// guestinfo.nightlight-router.mask
// guestinfo.nightlight-router.gw
// guestinfo.nightlight-router.dns1
// guestinfo.nightlight-router.dns2

func main() {

	ipaddr := "10.0.0.62"
	mask := ""
	gw := ""
	//dns1 := ""
	//dns2 := ""

	err := sysctl.Set("net.ipv4.ip_forward", "1")

	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("ROUTING VALUE: ", out)

	isVM, err := vmcheck.IsVirtualWorld(true)
	if err != nil {
		log.Printf("Error: %s", err)
	}

	if !isVM {
		log.Println("ERROR: not in a virtual world.")
	} else {
		log.Println("Running in VMWare")
		vsphereConfig := rpcvmx.NewConfig()
		ipaddr = vmtools.QueryGuestInfo(vsphereConfig, "guestinfo.nightlight-router.ip")
		mask = vmtools.QueryGuestInfo(vsphereConfig, "guestinfo.nightlight-router.mask")
		gw = vmtools.QueryGuestInfo(vsphereConfig, "guestinfo.nightlight-router.gw")
		//dns1 = vmtools.QueryGuestInfo(vsphereConfig, "guestinfo.nightlight-router.dns1")
		//dns2 = vmtools.QueryGuestInfo(vsphereConfig, "guestinfo.nightlight-router.dns2")
	}
	/*
		nics, err := netlink.LinkList()
		if err != nil {
			log.Println(err)
		}
	*/
	eth0, _ := netlink.LinkByName("eth0")
	ipAddress := ipaddr
	subnetMask := parseSubnetMask(mask)
	log.Println(ipAddress + subnetMask)
	addr, _ := netlink.ParseAddr(ipAddress + subnetMask)

	netlink.AddrReplace(eth0, addr)
	err = netlink.LinkSetUp(eth0)
	if err != nil {
		log.Fatalf("error bringing up the link: %v", err)
	}
	log.Println("Brought up NIC.")

	eth1, _ := netlink.LinkByName("eth1")
	eth1IpAddress := "192.168.10.1"
	eth1SubnetMask := parseSubnetMask("255.255.255.0")
	log.Println(eth1IpAddress + eth1SubnetMask)
	eth1addr, _ := netlink.ParseAddr(eth1IpAddress + eth1SubnetMask)

	netlink.AddrReplace(eth1, eth1addr)
	err = netlink.LinkSetUp(eth1)
	if err != nil {
		log.Fatalf("error bringing up the link: %v", err)
	}
	log.Println("Brought up NIC eth1")

	/*
		for _, nic := range nics {
			fmt.Println(nic.Attrs().Name)
			if nic.Attrs().Name == "eth1" {
				addr, _ := netlink.ParseAddr("192.168.10.1" + mask)
				netlink.AddrReplace(nic, addr)
				err = netlink.LinkSetUp(nic)
				if err != nil {
					log.Fatalf("error bringing up the link: %v", err)
				}
				log.Println("Brought up NIC eth1")
			}
		}
	*/
	defaultGateway := net.ParseIP(gw)

	defaultRoute := netlink.Route{
		Dst: nil,
		Gw:  defaultGateway,
	}

	if err := netlink.RouteAdd(&defaultRoute); err != nil {
		log.Fatal(err)
	}

	handler := cproxy.New(
		cproxy.Options.LogConnections(true),
		cproxy.Options.Logger(log.New(os.Stdout, "", log.LstdFlags)),
		//cproxy.Options.Filter(cproxy.NewHostnameFilter([]string{"google.com"})),
	)
	log.Println("Listening on:", "*:8080")
	_ = http.ListenAndServe(":8080", handler)

	//http.HandleFunc("/hello", hello)
	//http.ListenAndServe(":8090", nil)
}

func parseSubnetMask(mask string) (cidr string) {
	ip := net.ParseIP(mask)
	sz, _ := net.IPMask(ip.To4()).Size()
	return fmt.Sprintf("/%d", sz)
}

func hello(w http.ResponseWriter, req *http.Request) {
	nics, err := netlink.LinkList()
	if err != nil {
		log.Println(err)
	}
	var nicnames []string
	for _, nic := range nics {
		fmt.Println(nic.Attrs().Name)
		nicnames = append(nicnames, nic.Attrs().Name)
		nic.Attrs().HardwareAddr.String()
	}

	out := strings.Join(nicnames, ",")
	fmt.Fprintf(w, out)
}
