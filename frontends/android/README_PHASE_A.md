# Budget254 PAYE Android — Phase A

## What changed

- Premium calculator UI
- Guest calculations
- Historical date validation from 2022-01-01
- Custom deductions with validation
- Better network/API errors and 15-second timeout
- Loading and inline error states
- Improved mobile result screen
- API URL configured through `.env`

## Phone testing

1. Copy `.env.example` to `.env`.
2. Replace the sample IP with your laptop's Wi-Fi IPv4 address.
3. Start the Go API so it listens on `:8080`.
4. From the phone browser, open `http://YOUR_LAPTOP_IP:8080/health/live`.
5. Install dependencies and start Expo:

   npm install
   npx expo start --lan

6. Open Expo Go on the phone and scan the QR code.

Do not use `localhost` as the API address on a physical phone.
