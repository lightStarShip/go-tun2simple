package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
)

const (
	TLSHeaderLength = 5
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
func ParseHost(data []byte) string {
	switch data[0] {
	//GET,HEAD,POST,PUT,DELETE,OPTIONS,TRACE,CONNECT
	case 'G', 'H', 'P', 'D', 'O', 'T', 'C':
		{
			reader := bufio.NewReader(bytes.NewReader(data))
			if r, _ := http.ReadRequest(reader); r != nil {
				fmt.Sprintln("---===>>>Success host:", r.Host)
				return r.Host
			}
		}
	case 0x16:
		return getSNI(data)
	}
	return ""
}

func getSNI(data []byte) string {
	if len(data) == 0 || data[0] != 0x16 {
		return ""
	}

	extensions, err := GetExtensionBlock(data)
	if err != nil {
		return ""
	}
	sn, err := GetSNBlock(extensions)
	if err != nil {
		return ""
	}
	sni, err := GetSNIBlock(sn)
	if err != nil {
		return ""
	}

	fmt.Sprintln("---===>>>Success SNI:", string(sni))
	return string(sni)
}

func GetSNIBlock(data []byte) ([]byte, error) {
	index := 0
	if len(data) < 4 {
		return []byte{}, fmt.Errorf("not enough bytes to be an SNI block")
	}
	for {
		if index >= len(data) {
			break
		}
		length := int((data[index] << 8) + data[index+1])
		endIndex := index + 2 + length
		if data[index+2] == 0x00 { /* SNI */
			sni := data[index+3:]
			sniLength := int((sni[0] << 8) + sni[1])
			return sni[2 : sniLength+2], nil
		}
		index = endIndex
	}
	return []byte{}, fmt.Errorf(
		"finished parsing the SN block without finding an SNI",
	)
}
func GetExtensionBlock(data []byte) ([]byte, error) {
	/*   data[0]           - content type
	 *   data[1], data[2]  - major/minor version
	 *   data[3], data[4]  - total length
	 *   data[...38+5]     - start of SessionID (length bit)
	 *   data[38+5]        - length of SessionID
	 */
	var index = TLSHeaderLength + 38

	if len(data) <= index+1 {
		return []byte{}, fmt.Errorf("not enough bits to be a Client Hello")
	}

	/* Index is at SessionID Length bit */
	if newIndex := index + 1 + int(data[index]); (newIndex + 2) < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("not enough bytes for the SessionID")
	}

	/* Index is at Cipher List Length bits */
	if newIndex := index + 2 + int((data[index]<<8)+data[index+1]); (newIndex + 1) < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("not enough bytes for the Cipher List")
	}

	/* Index is now at the compression length bit */
	if newIndex := index + 1 + int(data[index]); newIndex < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("not enough bytes for the compression length")
	}

	/* Now we're at the Extension start */
	if len(data[index:]) == 0 {
		return nil, fmt.Errorf("no extensions")
	}
	return data[index:], nil
}

func GetSNBlock(data []byte) ([]byte, error) {
	index := 0

	if len(data) < 2 {
		return []byte{}, fmt.Errorf("not enough bytes to be an SN block")
	}

	extensionLength := int((data[index] << 8) + data[index+1])
	if extensionLength+2 > len(data) {
		return []byte{}, fmt.Errorf("extension looks bonkers")
	}
	data = data[2 : extensionLength+2]

	for {
		if index+3 >= len(data) {
			break
		}
		length := int((data[index+2] << 8) + data[index+3])
		endIndex := index + 4 + length
		if data[index] == 0x00 && data[index+1] == 0x00 {
			return data[index+4 : endIndex], nil
		}

		index = endIndex
	}

	return []byte{}, fmt.Errorf(
		"finished parsing the Extension block without finding an SN block",
	)
}
