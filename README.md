# kumpul-be

Backend service for Kumpul - a small event coordination app for friend groups.

## The Problem

Every week someone in the group has to organize badminton. That means proposing dates in the WhatsApp group, tallying votes manually, messaging the venue to check availability, and then chasing people for money afterward. It's the same painful loop every time, and someone always hasn't paid.

## What This Does

Kumpul handles the whole flow:

1. Creator makes an event with a few date + venue options
2. Everyone votes, either by opening the app (all events show up on the dashboard) or via a shareable link dropped in the group chat
3. Creator picks the winner, contacts the venue through a pre-filled WhatsApp message
4. People RSVP
5. Creator opens a payment session, the app splits the cost and tracks who's paid and who hasn't

## Stack

- **Go** - REST API
- **PostgreSQL** - main database
- **Redis** - sessions and rate limiting
- **Cloudinary** - payment proof image uploads
- **Google OAuth 2.0** - authentication

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 16
- Redis
- A Google OAuth app ([console.cloud.google.com](https://console.cloud.google.com))
- A Cloudinary account

### Setup

```bash
git clone https://github.com/yourname/kumpul-be
cd kumpul-be
cp config.yml.example config.yml
```

Fill in your `config.yml`, then:

```bash
# Run migrations
migrate -path db/migrations -database $DATABASE_URL up

# Start the server
go run main.go server
```

## Notes

- This is a personal project, not production-hardened
- WhatsApp integration is just pre-filled `wa.me` links, no API involved
- Payment is manual (transfer confirmation), no payment gateway
