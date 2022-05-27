# Shoutrrr ChatID Bot

Helper bot for getting the telegram Chat IDs for various chat types. 
Currently deployed on Google Cloud Platform using the name `@shoutrrrbot`.

## Deploy
```sh
# Set your preferred region (to whatever you want)
gcloud config set functions/region europe-west-1

# Add your bot token secret
echo "237895285:secretbottoken" | gcloud secrets create TelegramBotAPIKey

# Deploy function
gcloud functions deploy shoutrrr-chatid-bot \
    --set-env-vars "FUNCTION_REGION=$(gcloud config get functions/region),GCP_PROJECT=$(gcloud config get project)" \
    --set-secrets "BOT_API_TOKEN=TelegramBotAPIKey"
    --runtime go116 \
    --trigger-http \
    --allow-unauthenticated

# Register webhook (only needs to be done once)
curl $(gcloud functions describe shoutrrr-chatid-bot --format json | jq -r '.httpsTrigger.url + "?register"')

# ... or just visit the function deploy URL with `?register` appended  in the browser
```