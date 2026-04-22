'use strict'

const express = require('express')
const app = express()
const port = process.env.PORT || 8197

app.use(express.json())

app.get('/healthz', (_req, res) => res.json({ status: 'ok' }))
app.get('/metrics', (_req, res) => {
  res.set('Content-Type', 'text/plain')
  res.send('# placeholder metrics\n')
})

app.listen(port, () => console.log(`telegram-service listening on port ${port}`))
