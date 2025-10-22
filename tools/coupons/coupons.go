// coupons downloads two or more coupon sets and writes to stdout
// any coupons that appear in more than one set.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// usage: `go run ./tools/coupons1 -f "tmp/coupons/couponbase1 tmp/coupons/couponbase2 tmp/coupons/couponbase3"`
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, to io.Writer, args []string) error {
	set := flag.NewFlagSet("", flag.PanicOnError)
	files := set.String("f", "", "comma separated list of coupon files to check")
	set.Parse(args)

	fileNames := strings.Split(*files, " ")
	seen := map[string]int{}
	seenTwo := []string{}

	for _, fname := range fileNames {
		log.Println("processing", fname)
		r, err := os.Open(fname)
		if err != nil {
			return fmt.Errorf("failed to open coupon file: %w", err)
		}
		defer r.Close()

		s := bufio.NewScanner(r)
		for s.Scan() {
			if err := ctx.Err(); err != nil {
				return err
			}
			if len(s.Text()) < 8 || len(s.Text()) > 10 {
				// invalid coupon, ignore it
				continue
			}

			seen[s.Text()]++
			if seen[s.Text()] == 2 {
				seenTwo = append(seenTwo, s.Text())
			}
		}

		if s.Err() != nil {
			return fmt.Errorf("failed scanning coupon file: %w", err)
		}
	}

	for _, v := range seenTwo {
		if _, err := io.WriteString(to, v+"\n"); err != nil {
			return fmt.Errorf("failed to write seen: %w", err)
		}
	}
	return nil
}
