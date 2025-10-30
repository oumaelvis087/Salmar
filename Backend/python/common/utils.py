import json
import logging
import os
import time
from typing import Any, Dict, Optional

import dotenv

# Load environment variables
dotenv.load_dotenv()

# Configure logging
logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

def get_env_var(key: str, default: Optional[str] = None) -> str:
    """Get environment variable with fallback to default."""
    value = os.getenv(key, default)
    return value if value is not None else ""

def load_config() -> Dict[str, Any]:
    """Load configuration from environment variables."""
    return {
        "host": get_env_var("HOST", "0.0.0.0"),
        "port": int(get_env_var("PORT", "5000")),
        "debug": get_env_var("DEBUG", "False").lower() == "true",
        "model_path": get_env_var("MODEL_PATH", "./models"),
        "cache_dir": get_env_var("CACHE_DIR", "./cache"),
    }

def timer(func):
    """Decorator to measure function execution time."""
    def wrapper(*args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        execution_time = end_time - start_time
        logger.debug(f"Function {func.__name__} executed in {execution_time:.4f} seconds")
        return result
    return wrapper

def format_response(data: Dict[str, Any], processing_time: float = 0.0) -> Dict[str, Any]:
    """Format standard API response."""
    response = {
        **data,
        "processing_time": processing_time,
        "model_info": {
            "service": os.getenv("SERVICE_NAME", "unknown"),
            "version": os.getenv("SERVICE_VERSION", "0.1.0"),
        }
    }
    return response

def parse_request(request_data: Dict[str, Any]) -> Dict[str, Any]:
    """Parse and validate incoming request data."""
    # Basic validation could be added here
    return request_data