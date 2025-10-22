# Advanced Challenge

An implementation of the requirements at [backend-challenge/README.md](https://github.com/oolio-group/kart-challenge/blob/advanced-challenge/backend-challenge/README.md) aimed to give insight into my application building philosopy.

I like to design for simplicity and refactorbility with a strong focus on "Blackbox" tests to improve development velocity.

## Usage

run server: `make server`

run tests: `make test`

regenerate coupon database: `make api/coupons/data`

regenerate product database: `make api/products/data.json`

## Key Features

### Logging
Structured logging with opentelemtry trace and span information to aid debugging. Logs on meaningful events without being too noisy.

### OpenTelemetry Integration
Minimal example integration with OpenTelementry within the application for now exported nowhere and mostly used for unique request identifiers within logs.

### Separation of Concerns
The application is split into four main areas of concern:
- `cmd/server`: Application entrypoint. Parses CLI arguments and sets up dependency injection.
- `server`: HTTP layer. Handles wiring up business logic to routes, authn, logging, telemetry.
- `*`: Everything else. Business application and data access interfaces.

### Extensive Blackbox Tests
Fast, hermetic blackbox tests giving developers confidence that the application does what it is supposed to do. These use the applications public interface enabling refactoring with confidence.

## Decisions

### Embedded Coupon Stores
While the coupon source files are large (~1GB each when unzipped) the actual amount of duplicated coupons is quite small.

Given there are no requirements to update coupons on the fly and to keep this implementation simple and fit withing the expected time limit I've embedded the coupon codes within the server. The coupon code database is preprocessed to minimise size and application startup speed.

The use within the server has been kept generic and can easily be extended to use an external store such as an API or database with minimal impact to the code.

### Embedded Orders/Products Stores
Similarly to coupons, I've kept the implementations of the order and product data stores simple for this implementation. This left more time to focus on building a robost API framework.

The data stores have been integrated into the application in such a way that providing an alternative implementation such as a database would require minimal changes to the application.

## Extensions
Some easy areas for extension to turn this into a real live application:
- Add an external identity provider to create / provide certs to validate auth tokens.
- Depending on requirements for data mutability and availability replace the example memory in-memory / embedded data stores with an external store of some kind.
- Add an observability stack and update traces from main() to export there.
- Add a larger configuration source such as a config file. Needed as you start to add external dependencies.
- Add an itempotency key to orders to handle duplicate requests gracefully.
- Add extra request / response attributes of interest to [log](./monitoring/log.go). Some ideas include response status code and the request and response size in bytes.
