module.exports = function SSPError(message) {
    Error.captureStackTrace(this, this.constructor);
    this.name = 'SSPError';
    this.message = message;
}

require('util').inherits(module.exports, Error);