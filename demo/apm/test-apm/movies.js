require("./tracing");
const express = require('express');
const app = express()
const port = 3002
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

app.listen(port, () => {
    console.log(`Listening movies at http://localhost:${port}`)
    setTimeout(()=>{
        setInterval(()=>{
            http.get('http://localhost:3002/movies', (resp) => {
            }).on("error", (err) => {
                console.log("Error: " + err.message);
            });
        },1000)
    },2000)
})