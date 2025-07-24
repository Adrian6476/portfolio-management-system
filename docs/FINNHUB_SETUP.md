# Finnhub API Setup

This document explains how to securely configure the Finnhub API key for the portfolio management system.

## Local Development Setup

1. **Copy the environment template:**
   ```bash
   cp services/api-gateway/.env.example services/api-gateway/.env
   ```

2. **Add your Finnhub API key:**
   Edit `services/api-gateway/.env` and replace the placeholder:
   ```
   FINNHUB_API_KEY=your_actual_api_key_here
   ```

3. **Start the application:**
   ```bash
   docker compose up --build
   ```

## Features Enabled with Finnhub API

- **Real-time stock quotes:** Get current prices, daily changes, and market data
- **Company profiles:** Fetch real company names when adding new holdings
- **Enhanced portfolio insights:** Calculate real-time portfolio values

## API Usage

### Get Current Price
```bash
curl "http://localhost:8080/api/v1/market/price/AAPL"
```

### Add Holdings with Real Company Names
When you add a new holding, the system will automatically fetch the real company name from Finnhub:
```bash
curl -X POST "http://localhost:8080/api/v1/portfolio/holdings" \
  -H "Content-Type: application/json" \
  -d '{"symbol": "AAPL", "quantity": 10, "average_cost": 150.00}'
```

## GitHub Actions Setup

To run tests and deployments in GitHub Actions:

1. **Add the API key as a repository secret:**
   - Go to your repository on GitHub
   - Navigate to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `FINNHUB_API_KEY`
   - Value: Your actual Finnhub API key

2. **The workflow will automatically use the secret** for:
   - Backend tests that require market data
   - Building and deploying the application

## Rate Limits

- **Free tier:** 60 requests per minute
- **Recommended usage:** Implement caching for frequently requested data
- **Monitor usage:** Check your Finnhub dashboard regularly

## Security Notes

- ✅ The `.env` file is gitignored and won't be committed
- ✅ Docker Compose loads the environment variables securely
- ✅ GitHub Actions uses encrypted secrets
- ✅ The API key is only accessible to authorized environments

## Troubleshooting

### API Key Not Working
- Verify the key is correctly set in `.env`
- Check that Docker Compose is loading the environment file
- Ensure the key hasn't expired on the Finnhub dashboard

### Rate Limit Errors
- Implement caching to reduce API calls
- Consider upgrading to a paid Finnhub plan for higher limits
- Use batch requests when possible

## Without API Key

If no API key is provided:
- The application will still work for basic portfolio management
- Market data features will show "service unavailable" messages
- Company names will default to stock symbols
