var fyersModel = require('fyers-api-v3').fyersModel

var fyers = new fyersModel({
  path: './logs',
  enableLogging: true,
})

fyers.setAppId(process.env.FYERS_APP_ID)

fyers.setRedirectUrl(process.env.FYERS_REDIRECT_URL)

var URL = fyers.generateAuthCode()

//use url to generate auth code
console.log(URL)

var authcode = process.env.FYERS_AUTH_CODE
console.log(authcode, process.env.FYERS_APP_ID)
fyers
  .generate_access_token({
    client_id: process.env.FYERS_APP_ID,
    secret_key: process.env.FYERS_SECRET_KEY,
    auth_code: authcode,
  })
  .then((response) => {
    if (response.s == 'ok') {
      fyers.setAccessToken(response.access_token)
      console.log(response.access_token)
    } else {
      console.log('error generating access token', response)
    }
  })

// fyers
//   .get_profile()
//   .then((response) => {
//     console.log(response)
//   })
//   .catch((err) => {
//     console.log(err)
//   })

// fyers
//   .getQuotes(['NSE:SBIN-EQ', 'NSE:TCS-EQ'])
//   .then((response) => {
//     console.log(response)
//   })
//   .catch((err) => {
//     console.log(err)
//   })

// fyers
//   .getMarketDepth({ symbol: ['NSE:SBIN-EQ', 'NSE:TCS-EQ'], ohlcv_flag: 1 })
//   .then((response) => {
//     console.log(response)
//   })
//   .catch((err) => {
//     console.log(err)
//   })
