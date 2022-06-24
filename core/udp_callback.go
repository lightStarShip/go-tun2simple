package core

/*
#cgo CFLAGS: -I./lwip/src/include
#include "lwip/udp.h"

extern void udpRecvFn(void *arg, struct udp_pcb *pcb, struct pbuf *p, const ip_addr_t *addr, u16_t port);

void
set_udp_recv_callback(struct udp_pcb *pcb, void *recv_arg) {
	udp_recv(pcb, udpRecvFn, recv_arg);
}
*/
import "C"
import (
	"unsafe"
)

func setUDPRecvCallback(pcb *C.struct_udp_pcb, recvArg unsafe.Pointer) {
	C.set_udp_recv_callback(pcb, recvArg)
}

/*

typedef void (*udp_recv_fn)(void *arg, struct udp_pcb *pcb, struct pbuf *p,
    const ip_addr_t *addr, u16_t port);


*/
