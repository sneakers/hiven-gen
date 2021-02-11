# hiven-gen
Generate Hiven accounts from a wordlist

## How to use
Open `config.json` and put in your 2Captcha key, as well as your catchall.

Put all the usernames that you would like to check into `wordlist.txt`

Run the script: `go run main.go`

Your claimed usernames will be in `claimed.txt`

### Note
As of writing this, Hiven requires an invite code to create an account. If nothing changes on their end, then this will work when registrations are opened back up. You can still use this to check if usernames are available, however it will still try to register them. 
