from flask import Flask, request, jsonify
from flask_cors import CORS
import time
import sys
import os

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from common.utils import load_config, timer, format_response, parse_request
from image_service.processor import ImageProcessor

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Load configuration
config = load_config()
app.config.update(config)

# Initialize image processor
image_processor = ImageProcessor()

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint."""
    return jsonify({"status": "ok", "service": "image"})

@app.route('/process', methods=['POST'])
def process_request():
    """Process image requests."""
    try:
        data = request.json
        if not data:
            return jsonify({"error": "No data provided"}), 400
        
        with timer() as t:
            result = image_processor.process(data)
            
        return format_response(result, t.elapsed)
    except Exception as e:
        app.logger.error(f"Processing error: {str(e)}")
        return jsonify({"error": str(e)}), 500
        
@app.route('/multimodal', methods=['POST'])
def process_multimodal():
    """Process multimodal requests (image + text)."""
    try:
        data = request.json
        if not data or not data.get('image_url') or not data.get('text'):
            return jsonify({"error": "Missing required fields: image_url and text"}), 400
        
        with timer() as t:
            result = image_processor.process_multimodal(data)
            
        return format_response(result, t.elapsed)
    except Exception as e:
        app.logger.error(f"Multimodal processing error: {str(e)}")
        return jsonify({"error": str(e)}), 500

@app.route('/analyze', methods=['POST'])
def analyze_image():
    """Analyze an image."""
    start_time = time.time()
    
    try:
        data = request.get_json()
        if not data or not data.get('image_url'):
            return jsonify({"error": "No image URL provided"}), 400
        
        result = image_processor.analyze_image(data)
        processing_time = time.time() - start_time
        response = format_response(result, processing_time)
        return jsonify(response)
    except Exception as e:
        return jsonify({"error": f"Analysis error: {str(e)}"}), 500

@app.route('/generate', methods=['POST'])
def generate_image():
    """Generate an image based on text prompt."""
    start_time = time.time()
    
    try:
        data = request.get_json()
        if not data or not data.get('text'):
            return jsonify({"error": "No text prompt provided"}), 400
        
        result = image_processor.generate_image(data)
        processing_time = time.time() - start_time
        response = format_response(result, processing_time)
        return jsonify(response)
    except Exception as e:
        return jsonify({"error": f"Generation error: {str(e)}"}), 500

if __name__ == '__main__':
    port = int(os.environ.get('IMAGE_SERVICE_PORT', 5002))
    app.run(host='0.0.0.0', port=port, debug=config.get('debug', False))