const mongoose = require('mongoose');

mongoose.connect('mongodb://localhost:27017/grapher', {
  useNewUrlParser: true,
  useUnifiedTopology: true,
})
.then(() => console.log('MongoDB connected'))
.catch((err) => console.error(err));

module.exports = mongoose;