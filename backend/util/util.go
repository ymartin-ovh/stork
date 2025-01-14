package storkutil

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func UTCNow() time.Time {
	return time.Now().UTC()
}

// Returns URL of the host with port.
func HostWithPortURL(address string, port int64) string {
	return fmt.Sprintf("http://%s:%d/", address, port)
}

// Parses URL into host and port.
func ParseURL(url string) (host string, port int64) {
	ptrn := regexp.MustCompile(`https{0,1}:\/\/\[{1}(\S+)\]{1}(:([0-9]+)){0,1}`)
	m := ptrn.FindStringSubmatch(url)

	if len(m) == 0 {
		ptrn := regexp.MustCompile(`https{0,1}:\/\/([^\s\:\/]+)(:([0-9]+)){0,1}`)
		m = ptrn.FindStringSubmatch(url)
	}

	if len(m) > 1 {
		host = m[1]
	}

	if len(m) > 3 {
		p, err := strconv.Atoi(m[3])
		if err == nil {
			port = int64(p)
		}
	}
	return host, port
}

// Turns IP address into CIDR. If the IP address already seems to be using
// CIDR notation, it is returned.
func MakeCIDR(address string) (string, error) {
	if !strings.Contains(address, "/") {
		ip := net.ParseIP(address)
		if ip == nil {
			return address, errors.Errorf("provided string %s is not a valid IP address", address)
		}
		ip4 := ip.To4()
		if ip4 != nil {
			address += "/32"
		} else {
			address += "/128"
		}
	}
	return address, nil
}

// Recognizes if the specified value is an IP address or prefix and
// converts it if necessary. An IP address specified as 192.0.2.2/32
// will be converted to 192.0.2.2. Similarly, an IP address of
// 2001:db8:1::/128 will be converted to 2001:db8:1::. If the specified
// value is a prefix, e.g. 2001:db8:1::/48 it will be returned as is.
// The second returned value indicates if this is a prefix or address.
// The third value indicates if the specified value was valid and the
// conversion was successful.
func ParseIP(address string) (string, bool, bool) {
	ip, ipNet, err := net.ParseCIDR(address)
	if err != nil {
		return "", false, false
	}
	ones, bits := ipNet.Mask.Size()
	if ones == bits {
		return ip.String(), false, true
	}
	return ipNet.String(), true, true
}

// Formats provided string of hexadecimal digits to MAC address format
// using colon as separator. It returns formatted string and a boolean
// value indicating if the conversion was successful.
func FormatMACAddress(identifier string) (formatted string, ok bool) {
	// Check if the identifier is already in the desired format.
	identifier = strings.TrimSpace(identifier)
	pattern := regexp.MustCompile(`^[0-9A-Fa-f]{2}((:{1})[0-9A-Fa-f]{2})*$`)
	if pattern.MatchString(identifier) {
		// No conversion required. Return the input.
		return identifier, true
	}
	// We will have to convert it, but let's first check if this is a valid identifier.
	if !IsHexIdentifier(identifier) {
		return "", false
	}
	// Remove any colons and whitespaces.
	replacer := strings.NewReplacer(" ", "", ":", "")
	numericOnly := replacer.Replace(identifier)
	for i, character := range numericOnly {
		formatted += string(character)
		// Divide the string into groups with two digits.
		if i > 0 && i%2 != 0 && i < len(numericOnly)-1 {
			formatted += ":"
		}
	}
	return formatted, true
}

// Detects if the provided string is an identifier consisting of
// hexadecimal digits and optionally whitespace or colons between
// the groups of digits. For example: 010203, 01:02:03, 01::02::03,
// 01 02 03 etc. It is useful in detecting if the string comprises
// a DHCP client identifier or MAC address.
func IsHexIdentifier(text string) bool {
	pattern := regexp.MustCompile(`^[0-9A-Fa-f]{2}((\s*|:{0,2})[0-9A-Fa-f]{2})*$`)
	return pattern.MatchString(strings.TrimSpace(text))
}

func SetupLogging() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		// TODO: do more research and enable if it brings value
		// PadLevelText: true,
		// FieldMap: log.FieldMap{
		// 	FieldKeyTime:  "@timestamp",
		// 	FieldKeyLevel: "@level",
		// 	FieldKeyMsg:   "@message",
		// },
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// Grab filename and line of current frame and add it to log entry
			_, filename := path.Split(f.File)
			return "", fmt.Sprintf("%20v:%-5d", filename, f.Line)
		},
	})
}

// Helper code for mocking os/exec stuff... pathetic.
type Commander interface {
	Output(string, ...string) ([]byte, error)
}

type RealCommander struct{}

func (c RealCommander) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).Output()
}

// Convert bytes to hex string.
func BytesToHex(bytesArray []byte) string {
	var buf bytes.Buffer
	for _, f := range bytesArray {
		fmt.Fprintf(&buf, "%02X", f)
	}
	return buf.String()
}
