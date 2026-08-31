# Budget254 PAYE Android — Phase 6

React Native + Expo foundation consuming the Phase 5 Go API.

## Configure API
Create `.env`:

EXPO_PUBLIC_API_BASE_URL=http://YOUR-LAN-IP:8080

For Android Emulator, the default `http://10.0.2.2:8080` is used.

## Run
```bash
npm install
npx expo start
```

Then choose Android.

## Screens
- Calculator
- Calculation Result
- PAYE explanation

No login is required for basic calculations.

## Architecture
- `app/`: navigation/screens
- `components/`: reusable UI
- `lib/api.ts`: backend client
- `lib/types.ts`: API contracts
- `lib/money.ts`: display formatting

The frontend does not calculate Kenyan tax figures. It sends user inputs to the Go API and displays the versioned, date-specific result.
