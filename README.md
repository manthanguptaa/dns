# DNS Server & Resolver

This project is a DNS server and resolver implemented in the Go programming language. It provides the functionality to act as a DNS server, receiving DNS queries and responding with the appropriate DNS records. Additionally, it can function as a DNS resolver, allowing users to query external DNS servers for DNS records.

## Features

- DNS server: Accepts incoming DNS queries and responds with the appropriate DNS records.
- DNS resolver: Sends DNS queries to external DNS servers and receives the corresponding DNS records.

## Prerequisites

To use this project, you need to have the following installed:

- Go (version 1.19)
- Git

## Getting Started

Follow the steps below to get started with the DNS server and resolver:

1. Clone the repository using Git:

   ```bash
   git clone https://github.com/Manthan109/dns.git
   ```

2. Change to the project directory:

   ```bash
   cd dns
   ```

3. Run the project:

   ```bash
   go run cmd/dns-resolver/main.go
   ```

## Usage
Open a new tab in the terminal and type
```bash
dig @127.0.0.1 <url>
```
example
```bash
dig @127.0.0.1 google.com
```

## Project Notes
This is a simple DNS server and resolver that works only on UDP for now and IPv4 addresses (TypeA records). All the other missed out features like caching have been done intentionally to reduce the scope of the project.

## DNS Working
![Untitled-2023-06-05-2230](https://github.com/Manthan109/dns/assets/42516515/f9a106d2-7d5d-4b2e-8ab8-e6e4e0cecf49)

## Logical Flow
![logical-flow](https://github.com/Manthan109/dns/assets/42516515/43edf0bd-3692-4661-b70d-400ed32d7993)
