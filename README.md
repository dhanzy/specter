# Specter Security Scanner

A high-performance, plugin-based vulnerability scanning framework built in Go.
Specter uses YAML-driven templates to detect known CVEs, misconfigurations, and framework exposures across web applications and network services.

### Overview

Specter is designed to be:

- Fast (concurrent scanning engine)
- Extensible (YAML plugin architecture)
- Accurate (rule-based matcher engine)
- Framework aware (detects underlying technologies)
- Safe (detection-focused)

### it allows you to

- Scan a single domain or a list of domains
- Crawl and extract links
- Detect application frameworks
- Run CVE detection templates
- Generate structured results

### Architecture

```sh
Target → Framework Detection → Plugin Selection → HTTP/TCP Engine → Matcher Engine → Result
```

#### Core Components

- **Plugin Loader** – Loads YAML-based vulnerability templates
- **Framework Detector** – Identifies technologies (e.g., Next.js, WordPress, Express, etc.)
- **Execution Engine** – Sends crafted requests via HTTP/TCP
- **Matcher Engine** – Evaluates responses using regex, headers, status codes
- **Concurrent Worker Pool** – High-speed scanning across targets

### Installation

```sh
git clone https://github.com/dhanzy/specter.git
cd specter
make
```

### Usage

Scan and crawl from a seed target

```sh
./build/specter --target http://localhost:3000 -p cve-2025-55182
```

### Security Philosophy

Specter is designed for:

- Defensive security research
- Authorized penetration testing
- Framework exposure detection
- Vulnerability assessment workflows
- It focuses on detection rather than exploitation.
