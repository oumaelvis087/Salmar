from flask import Flask, request, jsonify
from flask_cors import CORS
import time
import sys
import os

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from common.utils import load_config, timer, format_response, parse_request
from nlp_service.processor import NLPProcessor

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Load configuration
config = load_config()
app.config.update(config)

# Initialize NLP processor
nlp_processor = NLPProcessor()

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint."""
    return jsonify({"status": "ok", "service": "nlp"})

@app.route('/classify_intent', methods=['POST'])
@timer
def classify_intent():
    """Classify the intent of a request."""
    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "No data provided"}), 400
        
        parsed_data = parse_request(data)
        intent = nlp_processor.classify_intent(parsed_data)
        
        return jsonify({"intent": intent})
    except Exception as e:
        return jsonify({"error": f"Intent classification error: {str(e)}"}), 500

@app.route('/process', methods=['POST'])
def process_text():
    """Process text input and return a response."""
    try:
        data = request.json
        if not data or 'text' not in data:
            return jsonify({"error": "Missing required field: text"}), 400
            
        text = data['text']
        history = data.get('history', [])
        intent = data.get('intent', None)  # Allow intent to be provided
        
        with timer() as t:
            result = nlp_processor.process(text, history, intent)
            
        return format_response(result, t.elapsed)
        
    except Exception as e:
        app.logger.error(f"Error processing text: {str(e)}")
        return jsonify({"error": str(e)}), 500
        
@app.route('/classify_intent', methods=['POST'])
def classify_intent():
    """Classify the intent of a text input."""
    try:
        data = request.json
        if not data or 'text' not in data:
            return jsonify({"error": "Missing required field: text"}), 400
            
        text = data['text']
        
        with timer() as t:
            intent = nlp_processor.classify_intent(text)
            
        result = {
            "intent": intent,
            "model_info": {
                "name": "Salmar Intent Classifier",
                "version": "0.1.0"
            }
        }
            
        return format_response(result, t.elapsed)
        
    except Exception as e:
        app.logger.error(f"Error classifying intent: {str(e)}")
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    port = int(os.environ.get('NLP_SERVICE_PORT', 5001))
    app.run(host='0.0.0.0', port=port, debug=config.get('debug', False))