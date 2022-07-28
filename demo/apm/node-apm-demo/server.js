const express = require('express');
const app = express()
const port = 3002

const tracker = require('@middlewarelabs-devs/melt-node-metrics')
tracker.track({
    MELT_API_KEY:'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjEsIkFjY291bnRJZCI6MSwiQXV0aFR5cGUiOiIiLCJUaW1lIjoiMjAyMi0wNi0zMFQwNjo0OTo0Ni4zODk2ODEzWiIsImlzcyI6Im13X19sb2dpbiIsInN1YiI6ImxvZ2luIn0.RlM4Zu0u-0lBvyUsVT2YRiPvWh-LeHNXv5bL0aAxuf0'
})

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
})