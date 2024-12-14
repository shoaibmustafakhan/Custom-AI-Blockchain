# Custom-AI-Blockchain
Customized scratch-made implementation of a Blockchain for simulation of data transactions, utilizing a deterministic Artificial Intelligence algorithm.

# Distributed Peer-to-Peer System with AI, IPFS, and Blockchain

## Overview
This system connects three peers in a decentralized network and performs various operations on transaction data. The system features an AI-driven deterministic algorithm, supports peer-to-peer data exchange with flooding/gossiping protocols, and integrates IPFS (InterPlanetary File System) for distributed file storage. It also includes blockchain functionality with block mining and proof of work (crypto puzzle).

## Features
- **Connecting 3 Peers:** The system connects three peers in a decentralized network.
- **Transaction Handling:** Each peer processes more than 5 transactions with deterministic AI algorithms applied to the data.
- **Flooding/Gossiping Protocol:** Data is disseminated across all peers using flooding/gossiping for efficient and robust communication.
- **Blockchain:** Each peer participates in mining blocks, and the network ensures consensus through proof of work (crypto puzzle).
- **IPFS Integration:** Distributed file storage across peers using IPFS.

## Installation

### Prerequisites
1. **Go:** Ensure you have Go installed on your machine. If not, install it from [here](https://golang.org/dl/).
2. **IPFS:** Follow the steps below to install IPFS on Windows.

### IPFS Installation on Windows
1. **Download IPFS:**
   - Visit the IPFS download page: [https://dist.ipfs.tech/#kubo](https://dist.ipfs.tech/#kubo).
   - Download the Windows binary (e.g., `kubo_v0.25.0_windows-amd64.zip`).
2. **Extract the Files:**
   - Extract the zip file to a directory on your system.
3. **Update PATH Environment Variable:**
   - Add the extracted folder to your system's PATH environment variable.
4. **Initialize IPFS:**
   - Open Command Prompt and run the following commands:
     ```bash
     ipfs init
     ipfs daemon
     ```

### Go IPFS Dependency
To add the IPFS dependency to your Go project, run the following command in Command Prompt:
```bash
go get github.com/ipfs/go-ipfs-api
