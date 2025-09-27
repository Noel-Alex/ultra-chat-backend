<p align="center">
<a href="https://dscvit.com">
	<img width="400" src="https://user-images.githubusercontent.com/56252312/159312411-58410727-3933-4224-b43e-4e9b627838a3.png#gh-light-mode-only" alt="GDSC VIT"/>
</a>
	<h2 align="center"> Ultra Chat Backend </h2>
	<h4 align="center"> Backend written in golang for ultra-chat-bot <h4>
</p>

---

[![Join Us](https://img.shields.io/badge/Join%20Us-Developer%20Student%20Clubs-red)](https://dsc.community.dev/vellore-institute-of-technology/)
[![Discord Chat](https://img.shields.io/discord/760928671698649098.svg)](https://discord.gg/498KVdSKWR)

[![DOCS](https://img.shields.io/badge/Documentation-see%20docs-green?style=flat-square&logo=appveyor)](INSERT_LINK_FOR_DOCS_HERE)
[![UI ](https://img.shields.io/badge/User%20Interface-Link%20to%20UI-orange?style=flat-square&logo=appveyor)](INSERT_UI_LINK_HERE)

## Table of Contents

- [Key Features](#key-features)
  - [Authentication](#authentication)
  - [Summaries](#summaries)
- [Usage](#usage)
  - [Installation](#installation)
  - [Running](#running)
- [API Documentation](#api-documentation)
- [Developer](#developer)

<br>

## Key Features

### Authentication

- [x] Discord OAuth2 Integration
- [x] JWT based authentication
- [x] User profile management

### Summaries

- [x] Create chat summaries
- [x] Retrieve user summaries
- [x] Update existing summaries
- [x] Delete summaries
- [x] Private/Public summary options

<br>

## Usage

### Installation

```bash
# Clone the repository
git clone https://github.com/GDGVIT/ultra-chat-backend.git

# Install dependencies
go mod download
```

### Running

```bash
# Set environment variables
export PORT=5001
export MONGODB_URI=your_mongodb_uri
export CLIENT_ID=your_discord_client_id
export CLIENT_SECRET=your_discord_client_secret
export REDIRECT_URI=your_redirect_uri

# Run the server
go run main.go
```

## API Documentation

Postman API Documentations - [Ultra Chat Backend](https://github.com/DevloperAmanSingh/ultra-chat-backend/blob/main/postman.json)

Key Endpoints:

- POST /create-summary - Create a new chat summary
- GET /summarizer - Get user summaries
- PUT /update-summary - Update existing summary
- DELETE /delete-summary - Delete a summary
- GET /is_authenticated - Check authentication status

<br>

## Developer

<table>
	<tr align="center">
		<td>
		Aman Singh
		<p align="center">
			<img src = "https://avatars.githubusercontent.com/u/80804989?v=4" width="150" height="150" alt="Aman Singh">
		</p>
			<p align="center">
				<a href = "https://github.com/DevloperAmanSingh">
					<img src = "http://www.iconninja.com/files/241/825/211/round-collaboration-social-github-code-circle-network-icon.svg" width="36" height = "36" alt="GitHub"/>
				</a>
				<a href = "https://www.linkedin.com/in/amansingh2112">
					<img src = "http://www.iconninja.com/files/863/607/751/network-linkedin-social-connection-circular-circle-media-icon.svg" width="36" height="36" alt="LinkedIn"/>
				</a>
			</p>
		</td>
	</tr>
<td>
		Noel Alex
		<p align="center">
			<img src = "https://avatars.githubusercontent.com/u/79050483?v=4" width="150" height="150">
		</p>
			<p align="center">
				<a href = "https://github.com/Noel-alex“>
					<img src = "http://www.iconninja.com/files/241/825/211/round-collaboration-social-github-code-circle-network-icon.svg" width="36" height = "36" alt="GitHub"/>
				</a>
				<a href = "https://www.linkedin.com/in/noel-alex-b1731128b/”>
					<img src = "http://www.iconninja.com/files/863/607/751/network-linkedin-social-connection-circular-circle-media-icon.svg" width="36" height="36" alt="LinkedIn"/>
				</a>
			</p>
		</td>
</table>

<p align="center">
	Made with ❤ by <a href="https://dscvit.com">GDSC-VIT</a>
</p>
