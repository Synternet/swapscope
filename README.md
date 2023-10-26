# SwapScope
[![Latest release](https://img.shields.io/github/v/release/SyntropyNet/swapscope)](https://github.com/SyntropyNet/swapscope/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/SyntropyNet/swapscope/github-ci.yml?label=github-ci)](https://github.com/SyntropyNet/swapscope/actions/workflows/github-ci.yml)
[![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/SyntropyNet/swapscope/docker-image.yml?label=docker-image)](https://github.com/SyntropyNet/swapscope/actions/workflows/docker-image.yml)

Swapscope is an open-source project that simplifies the process of streaming real-time Ethereum Uniswap (for now) liquidity operations (additions and removals) data. The project consists of two main components: the DApp (frontend) and the publisher (backend). 
Swapscope's publisher utilizes [Syntropy Data Layer's (DL)](https://www.syntropynet.com/post/presenting-the-new-vision) Ethereum event log stream, processes the data, and publishes it back to Syntropy DL which frontend also utilizes to display the processed data.

## Getting Started

### Prerequisites
Before you start using (modifying) Swapscope locally, there are some prerequisites you need to fulfill:
* Access to Syntropy Data Layer and its streams - [Developer Portal](https://developer-portal.syntropynet.com/)
* Go (Golang 1.20) - [Install Go](https://go.dev/doc/install)
* Node.js and npm - [Install Node.js and npm](https://nodejs.org/en)
* Access to Coingecko API (for token and pricing information (for now)) - [Coingecko API](https://www.coingecko.com/en/api)
* Local (Docker) PostgreSQL database for storing token and liquidity pool information (helps to reduce API calls significantly) + publisher has built-in functionality to save liquidity addition/removal operations into a database

### Usage (individual)
* Latest release images can be found here: https://github.com/SyntropyNet/swapscope/releases/latest
* DApp instructions: [DApp README.md](dapp/README.md)
* Publisher instructions: [Publisher README.md](publisher/README.md)

### QuickStart (Docker compose)

❗ Requires dapp/.env
❗ As of 2023-10-26 dapp/.env file has to be expanded with `NEXT_PUBLIC_ENV` variable and value `"development"`.
❗ Requires publisher .env file in the root directory

1. Build.
```
docker-compose build
```

2. Start.
```
docker-compose up
```
&emsp;&emsp;or
```
docker-compose start
```

## Future development
As this is a young project, there still is a lot of room to improve! Some key features that are planned for the future:
* Swaps processing
* More Decentralised Exchanges:
  * PancakeSwap
  * SushiSwap
* More statistics and processed data:
  * Consolidated information about liquidity providers
  * Liquidity position earned fees and profitability

## Contributing
We welcome contributions from the community. Whether it's a bug report, a new feature, or a code fix, your input is valued and appreciated.

## Syntropy
If you have any questions, ideas, or simply want to connect with us, we encourage you to reach out through any of the following channels:

- **Discord**: Join our vibrant community on Discord at [https://discord.com/invite/jqZur5S3KZ](https://discord.com/invite/jqZur5S3KZ). Engage in discussions, seek assistance, and collaborate with like-minded individuals.
- **Telegram**: Connect with us on Telegram at [https://t.me/SyntropyNet](https://t.me/SyntropyNet). Stay updated with the latest news, announcements, and interact with our team members and community.
- **Email**: If you prefer email communication, feel free to reach out to us at devrel@syntropynet.com. We're here to address your inquiries, provide support, and explore collaboration opportunities.

## License
This project is licensed under the terms of the MIT license.