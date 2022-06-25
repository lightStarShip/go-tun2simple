package core

/*
#cgo CFLAGS: -I./lwip/src/include
#include "lwip/tcp.h"

extern err_t output(struct pbuf *p);

err_t
output_ip4(struct netif *netif, struct pbuf *p, const ip4_addr_t *ipaddr)
{
	return output(p);
}

err_t
output_ip6(struct netif *netif, struct pbuf *p, const ip6_addr_t *ipaddr)
{
	return output(p);
}

void
set_output()
{
	if (netif_list != NULL) {
		(*netif_list).output = output_ip4;
		(*netif_list).output_ip6 = output_ip6;
	}
}
*/
import "C"

func init() {
	C.set_output()
}
