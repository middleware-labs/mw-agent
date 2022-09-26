require("./tracing");
require("./fluent")
const express = require('express');
const app = express()
const port = 3003
const http = require('http');


app.get('/movies', async function (req, res) {
    res.type('json')
    setTimeout((() => {
        res.send(({
            movies: [
                {name: 'Jaws', genre: 'Thriller'},
                {name: 'Annie', genre: 'Family'},
                {name: 'Jurassic Park', genre: 'Action'},
            ]
        }))
    }), 1000)
})

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
