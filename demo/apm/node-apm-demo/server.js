const express = require('express');
const app = express()
const port = 3002
const cors = require("cors");

app.use(cors({origin: `http://localhost:${port}`}));

app.use(express.json()); /* bodyParser.json() is deprecated */

app.use(express.urlencoded({ extended: true })); /* bodyParser.urlencoded() is deprecated */

app.get('/500-error', async function (req, res) {
    return res.status(500).send('Internal Server Occurred');
})

app.get('/504-error', async function (req, res) {
    return res.status(504).send('Timeout error');
})

app.listen(port, () => {
    console.log(`Listening movies at http://localhost:${port}`)
})

require("./app/routes/tutorial.routes.js")(app);

const tracker = require('@middlewarelabs-devs/melt-node-metrics')

tracker.track({
    MELT_API_KEY:'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjEsIkFjY291bnRJZCI6MSwiQXV0aFR5cGUiOiIiLCJUaW1lIjoiMjAyMi0wOC0wOFQxMDoyMjo0My4yNTQxODRaIiwiaXNzIjoibXdfX2xvZ2luIiwic3ViIjoibG9naW4ifQ._kD_wnP6WKaHYq9VHaEpiEiktKS7hjRCiui4OveAWgE'
})

