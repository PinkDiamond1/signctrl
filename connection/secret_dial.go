package connection

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

var (
	// ErrAbortDial is returned if either SIGINT or SIGTERM are fired into the quit
	// channel.
	ErrAbortDial = errors.New("dialing aborted")
)

// RetrySecretDialTCP keeps dialing the given TCP address until success, using the
// given privkey for encryption and returns the secret connection.
func RetrySecretDialTCP(address string, privkey tm_ed25519.PrivKey, logger *log.Logger) (net.Conn, error) {
	logger.Printf("[INFO] signctrl: Dialing %v... (Use Ctrl+C to abort)", address)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// RetryDialInterval is the interval in which SignCTRL tries to repeatedly dial
	// the validator.
	// Make SignCTRL dial immediately the first time.
	RetryDialInterval := time.Duration(0)

	for {
		select {
		case <-sigs:
			return nil, ErrAbortDial

		case <-time.After(RetryDialInterval):
			if conn, err := net.Dial("tcp", strings.TrimPrefix(address, "tcp://")); err == nil {
				logger.Println("[INFO] signctrl: Successfully dialed the validator ✓")
				return tm_p2pconn.MakeSecretConnection(conn, privkey)
			}

			// After the first dial, dial in intervals of 1 second.
			RetryDialInterval = time.Second
			logger.Println("[DEBUG] signctrl: Retry dialing...")
		}
	}
}
