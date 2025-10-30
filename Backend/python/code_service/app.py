from flask import Flask, request, jsonify
from flask_cors import CORS
import time
import sys
import os

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from common.utils import load_config, timer, format_response, parse_request
from code_service.processor import CodeProcessor

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Load configuration
config = load_config()
app.config.update(config)

# Initialize code processor
code_processor = CodeProcessor()

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint."""
    return jsonify({"status": "ok", "service": "code"})

@app.route('/process', methods=['POST'])
def process_request():
    """Process code-related requests."""
    try:
        data = request.json
        if not data:
            return jsonify({"error": "No data provided"}), 400
        
        with timer() as t:
            result = code_processor.process(data)
            
        return format_response(result, t.elapsed)
    except Exception as e:
        app.logger.error(f"Processing error: {str(e)}")
        return jsonify({"error": str(e)}), 500
        
@app.route('/multimodal/code-text', methods=['POST'])
def process_code_with_text():
    """Process multimodal requests (code + text)."""
    try:
        data = request.json
        if not data or not data.get('code') or not data.get('text'):
            return jsonify({"error": "Missing required fields: code and text"}), 400
        
        with timer() as t:
            result = code_processor.process_multimodal(data)
            
        return format_response(result, t.elapsed)
    except Exception as e:
        app.logger.error(f"Multimodal processing error: {str(e)}")
        return jsonify({"error": str(e)}), 500
        
@app.route('/multimodal/code-image', methods=['POST'])
def process_code_with_image():
    """Process multimodal requests (code + image)."""
    try:
        data = request.json
        if not data or not data.get('code') or not data.get('image_url'):
            return jsonify({"error": "Missing required fields: code and image_url"}), 400
        
        with timer() as t:
            result = code_processor.process_code_with_image(data)
            
        return format_response(result, t.elapsed)
    except Exception as e:
        app.logger.error(f"Multimodal processing error: {str(e)}")
        return jsonify({"error": str(e)}), 500

@app.route('/generate', methods=['POST'])
def generate_code():
    """Generate code based on text description."""
    start_time = time.time()
    
    try:
        data = request.get_json()
        if not data or not data.get('text'):
            return jsonify({"error": "No text description provided"}), 400
        
        result = code_processor.generate_code(data)
        processing_time = time.time() - start_time
        response = format_response(result, processing_time)
        return jsonify(response)
    except Exception as e:
        return jsonify({"error": f"Generation error: {str(e)}"}), 500

@app.route('/explain', methods=['POST'])
def explain_code():
    """Explain provided code."""
    start_time = time.time()
    
    try:
        data = request.get_json()
        if not data or not data.get('code'):
            return jsonify({"error": "No code provided"}), 400
        
        result = code_processor.explain_code(data)
        processing_time = time.time() - start_time
        response = format_response(result, processing_time)
        return jsonify(response)
    except Exception as e:
        return jsonify({"error": f"Explanation error: {str(e)}"}), 500

if __name__ == '__main__':
    port = int(os.environ.get('CODE_SERVICE_PORT', 5003))
    app.run(host='0.0.0.0', port=port, debug=config.get('debug', False))