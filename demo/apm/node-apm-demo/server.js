const tracker = require('@middlewarelabs-devs/mw-node-apm')
tracker.track()

const express = require('express');
const app = express()
const port = 3002
const cors = require("cors");

app.use(cors({origin: `http://localhost:${port}`}));

app.use(express.json()); /* bodyParser.json() is deprecated */

app.use(express.urlencoded({ extended: true })); /* bodyParser.urlencoded() is deprecated */

app.get('/304-error', async function (req, res) {
    return res.status(304).send('Not Modified');
})

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
if (process.env.MW_AUTOGENERATE_TRACING_DATA) {
    require("./tracingloop")
}

if (!process.env.MW_APM_PAUSE_LOGS) {
    require("./fluent")
    require("./fluent-winston")
}
