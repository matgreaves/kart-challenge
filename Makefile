# run application
.PHONY: server
server:
	go run ./api/cmd/server

# run tests
.PHONY: test
test:
	go test ./...

# create embedded product database
.PHONY: api/products/data.json
api/products/data.json:
	curl https://orderfoodonline.deno.dev/api/product > $@

# create embedded coupon database
.PHONY: api/coupons/data
api/coupons/data: tmp/coupons/couponbase1 tmp/coupons/couponbase2 tmp/coupons/couponbase3
	go run ./tools/coupons -f "$^" > $@

# compile the application
.PHONY: bin
bin:
	mkdir -p tmp/bin
	CGO_ENABLED=0 go build -o ./tmp/bin/kart ./api/cmd/server

# build app into a Docker container
.PHONY: docker
docker:
	docker build . -t kart

# downloand coupon source files
tmp/coupons/%:
	mkdir -p tmp/coupons
	curl https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/$*.gz | gunzip > tmp/coupons/$*

# removed generated and cached content
.PHONY: clean
clean:
	rm -rf tmp
	rm api/products/data.json
	rm api/order/data
