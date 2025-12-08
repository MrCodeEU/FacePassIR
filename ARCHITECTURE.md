# FacePass Technical Architecture

## System Overview

```
┌─────────────┐
│   User      │
│  Attempts   │
│    Login    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│              PAM (Pluggable Authentication)         │
│  ┌──────────────────────────────────────────────┐   │
│  │  auth    required   pam_facepass.so          │   │
│  │  auth    requisite  pam_unix.so              │   │
│  └──────────────────────────────────────────────┘   │
└──────────────────┬──────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────┐
│              FacePass-PAM Binary                    │
│  ┌──────────────────────────────────────────────┐   │
│  │  1. Read username from PAM                   │   │
│  │  2. Initialize camera with IR emitter        │   │
│  │  3. Capture frames (multi-angle)             │   │
│  │  4. Perform liveness detection               │   │
│  │  5. Extract face embeddings                  │   │
│  │  6. Compare with stored embeddings           │   │
│  │  7. Return success/failure to PAM            │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
           │                    │
           │ Success            │ Failure/Timeout
           ▼                    ▼
    ┌───────────┐        ┌──────────────┐
    │  Unlock   │        │   Password   │
    │  Desktop  │        │   Fallback   │
    └───────────┘        └──────────────┘
```

---

## Component Architecture

### 1. Camera Module (`pkg/camera/`)

**Purpose:** Interface with camera hardware, especially IR cameras

```go
type Camera interface {
    Open(device string) error
    Close() error
    Capture() (Frame, error)
    GetDeviceInfo() DeviceInfo
}

type IRCamera struct {
    device     string
    emitter    *IREmitter
    handle     *v4l2.Device
    isOpen     bool
}

// Frame represents a single camera frame
type Frame struct {
    Data      []byte
    Width     int
    Height    int
    Format    string    // "RGB", "GRAY", "IR"
    Timestamp time.Time
}
```

**Data Flow:**
```
Camera Device → V4L2 Driver → Frame Buffer → Go []byte → Frame struct
```

---

### 2. Recognition Module (`pkg/recognition/`)

**Purpose:** Face detection, recognition, and embedding generation

```go
type Recognizer interface {
    LoadModels(path string) error
    DetectFaces(frame Frame) ([]Face, error)
    RecognizeFace(face Face) (Embedding, error)
    CompareFaces(emb1, emb2 Embedding) (float64, error)
}

type Face struct {
    BoundingBox Rectangle
    Landmarks   []Point      // Eye corners, nose, mouth corners
    Confidence  float64
}

type Embedding struct {
    Vector []float32  // 128-dimensional face embedding
    Quality float64   // Confidence score
}
```

**Data Flow:**
```
Frame → Face Detection → Face Crop → Alignment →
Feature Extraction → 128D Embedding → Comparison
```

**Technical Details:**
- Uses dlib's ResNet face recognition model
- 128-dimensional embeddings
- Euclidean distance for comparison
- Threshold typically 0.6 (configurable)

---

### 3. Liveness Detection Module (`pkg/liveness/`)

**Purpose:** Prevent spoofing attacks (photos, videos, masks)

```go
type LivenessDetector interface {
    Detect(frames []Frame, config LivenessConfig) LivenessResult
}

type LivenessConfig struct {
    RequireBlink       bool
    RequireConsistency bool
    RequireChallenge   bool
    EnableIRAnalysis   bool
    MinScore           float64
}

type LivenessResult struct {
    IsLive      bool
    Score       float64
    Checks      map[string]bool  // "blink", "consistency", etc.
    Reason      string           // If failed, why
}
```

**Multi-Tier Detection:**

**Tier 1: Basic (Always On)**
```
Frames (5-10 over 2s) → Blink Detection + Consistency Check → Basic Score
```

**Tier 2: Challenge-Response**
```
Random Challenge → User Response → Movement Analysis → Enhanced Score
```

**Tier 3: IR Analysis (If Available)**
```
IR Frame → Reflection Pattern Analysis + Texture Analysis → Final Score
```

---

### 4. Storage Module (`pkg/storage/`)

**Purpose:** Securely store and retrieve face embeddings

```go
type Storage interface {
    SaveUser(user UserFaceData) error
    LoadUser(username string) (UserFaceData, error)
    DeleteUser(username string) error
    ListUsers() ([]string, error)
}

type UserFaceData struct {
    Username    string
    Embeddings  []Embedding  // Multiple angles
    EnrolledAt  time.Time
    LastUsed    time.Time
    Metadata    map[string]string
}
```

**File Structure:**
```
~/.local/share/facepass/
├── config.yaml
├── users/
│   ├── john.json.enc     # Encrypted embeddings
│   ├── jane.json.enc
│   └── alice.json.enc
├── models/
│   ├── dlib_face_recognition_resnet_model_v1.dat
│   └── shape_predictor_5_face_landmarks.dat
└── logs/
    └── facepass.log
```

**Encryption:**
- Uses NaCl secretbox (XSalsa20 + Poly1305)
- Key derived from system information (CPU ID, MAC address)
- Prevents embedding theft/reuse on different machines

**Example Encrypted Storage:**
```json
{
  "username": "john",
  "embeddings": [
    {
      "vector": "[encrypted base64]",
      "angle": "front",
      "quality": 0.95
    },
    {
      "vector": "[encrypted base64]",
      "angle": "left",
      "quality": 0.92
    }
  ],
  "enrolled_at": "2025-12-08T10:30:00Z",
  "metadata": {
    "camera": "IR",
    "version": "0.1.0"
  }
}
```

---

### 5. PAM Integration Module (`pkg/pam/`)

**Purpose:** Interface with Linux PAM system

```go
type PAMAuth struct {
    username string
    config   *config.Config
    logger   *logrus.Logger
}

func (p *PAMAuth) Authenticate() error {
    // 1. Load user's face data
    userData, err := storage.LoadUser(p.username)

    // 2. Initialize camera
    camera := camera.NewIRCamera()
    camera.Open()
    defer camera.Close()

    // 3. Capture frames
    frames := captureMultipleFrames(camera, 10)

    // 4. Liveness detection
    liveness := livenessDetector.Detect(frames, livenessConfig)
    if !liveness.IsLive {
        return fmt.Errorf("liveness check failed: %s", liveness.Reason)
    }

    // 5. Face recognition
    embedding := recognizer.RecognizeFace(frames[0])

    // 6. Compare with stored embeddings
    bestMatch := findBestMatch(embedding, userData.Embeddings)

    if bestMatch.Distance < threshold {
        return nil  // Success
    }

    return errors.New("face not recognized")
}
```

**PAM Configuration (`/etc/pam.d/common-auth`):**
```
# FacePass authentication
auth    [success=2 default=ignore]  pam_exec.so quiet /usr/local/bin/facepass-pam
auth    [success=1 default=ignore]  pam_unix.so nullok_secure try_first_pass
auth    requisite                   pam_deny.so
auth    required                    pam_permit.so
```

---

## Authentication Flow (Detailed)

### Enrollment Flow
```
1. User runs: facepass enroll username
                    ↓
2. Detect IR camera and enable emitter
                    ↓
3. Show instructions: "Look at camera..."
                    ↓
4. Capture 5-7 angles (front, left, right, up, down)
                    ↓
5. For each angle:
   - Detect face
   - Extract landmarks
   - Generate 128D embedding
   - Verify quality (reject if poor)
                    ↓
6. Store embeddings (encrypted)
                    ↓
7. Test recognition immediately
                    ↓
8. Success message + tips
```

### Authentication Flow
```
1. PAM calls facepass-pam with username
                    ↓
2. Load user's face data from storage
                    ↓
3. Initialize camera (enable IR if available)
                    ↓
4. Start timeout timer (default 10s)
                    ↓
5. Capture frames continuously
                    ↓
6. PARALLEL PROCESSING:
   ├─→ Liveness Detection (Tier 1)
   │   ├─ Blink detection
   │   └─ Frame consistency
   │
   ├─→ Face Detection & Recognition
   │   ├─ Detect face in frame
   │   ├─ Extract embedding
   │   └─ Compare with stored
   │
   └─→ IR Analysis (if available)
       ├─ Reflection patterns
       └─ Texture analysis
                    ↓
7. Calculate combined score
                    ↓
8. Decision:
   - Score > threshold → SUCCESS → Unlock
   - Score < threshold → Continue capture
   - Timeout reached  → FAILURE → Password fallback
```

---

## Security Considerations

### 1. Embedding Security
- **Encrypted at rest:** Uses NaCl secretbox
- **Machine-specific key:** Cannot transfer to another system
- **No reversibility:** Embeddings cannot be converted back to images

### 2. Anti-Spoofing (Multi-Layer)
```
┌─────────────────────────────────────────┐
│  Layer 1: Blink Detection               │  ← Defeats photos
├─────────────────────────────────────────┤
│  Layer 2: Multi-frame Consistency       │  ← Defeats static images
├─────────────────────────────────────────┤
│  Layer 3: Challenge-Response            │  ← Defeats videos
├─────────────────────────────────────────┤
│  Layer 4: IR Reflection Analysis        │  ← Defeats screens
├─────────────────────────────────────────┤
│  Layer 5: Texture Analysis              │  ← Defeats printouts
└─────────────────────────────────────────┘
```

### 3. Privacy Protection
- **No image storage:** Only mathematical embeddings stored
- **IR images:** Less identifiable than RGB photos
- **Local processing:** No cloud/network required
- **User control:** Easy enrollment/removal

### 4. Fail-Safe Design
- **Always fallback to password:** Never lock out users
- **Timeout protection:** Won't hang indefinitely
- **Graceful degradation:** Works without IR, without liveness detection
- **Clear error messages:** Users know why authentication failed

---

## Performance Targets

| Operation              | Target Time | Acceptable Range |
|------------------------|-------------|------------------|
| Face Detection         | 50ms        | 30-100ms         |
| Embedding Generation   | 100ms       | 50-200ms         |
| Embedding Comparison   | 1ms         | <5ms             |
| Liveness Detection     | 500ms       | 200ms-1s         |
| Total Authentication   | 1-2s        | <3s              |
| Enrollment (5 angles)  | 10s         | 5-15s            |

---

## Configuration Profiles

### Basic (Fast, Less Secure)
```yaml
liveness_detection:
  level: basic
  blink_required: true
  consistency_check: true
  challenge_response: false
  min_liveness_score: 0.5

recognition:
  confidence_threshold: 0.7  # More permissive
```

### Standard (Balanced)
```yaml
liveness_detection:
  level: standard
  blink_required: true
  consistency_check: true
  challenge_response: false
  min_liveness_score: 0.7

recognition:
  confidence_threshold: 0.6
```

### Strict (Secure, Slower)
```yaml
liveness_detection:
  level: strict
  blink_required: true
  consistency_check: true
  challenge_response: true
  ir_analysis: true
  min_liveness_score: 0.8

recognition:
  confidence_threshold: 0.5  # More strict
```

### Paranoid (Maximum Security)
```yaml
liveness_detection:
  level: paranoid
  blink_required: true
  consistency_check: true
  challenge_response: true
  ir_analysis: true
  texture_analysis: true
  min_liveness_score: 0.9

recognition:
  confidence_threshold: 0.4  # Very strict

auth:
  require_manual_review: true  # Flag for manual check
```

---

## Error Handling Strategy

```go
type AuthError struct {
    Code    ErrorCode
    Message string
    Retry   bool
    Details map[string]interface{}
}

const (
    ErrNoFaceDetected     ErrorCode = "NO_FACE"
    ErrMultipleFaces      ErrorCode = "MULTIPLE_FACES"
    ErrLivenessFailure    ErrorCode = "LIVENESS_FAILED"
    ErrRecognitionFailure ErrorCode = "NOT_RECOGNIZED"
    ErrCameraFailure      ErrorCode = "CAMERA_ERROR"
    ErrTimeout            ErrorCode = "TIMEOUT"
    ErrNoEnrollment       ErrorCode = "NOT_ENROLLED"
)
```

**User-Friendly Messages:**
- `NO_FACE` → "Please position your face in front of the camera"
- `MULTIPLE_FACES` → "Multiple faces detected. Please ensure only you are in frame"
- `LIVENESS_FAILED` → "Liveness check failed. Please blink and try again"
- `NOT_RECOGNIZED` → "Face not recognized. Falling back to password..."
- `TIMEOUT` → "Face recognition timed out. Please enter your password"

---

## Testing Strategy

### Unit Tests
- Camera frame capture
- Face detection accuracy
- Embedding comparison
- Encryption/decryption
- Liveness detection algorithms

### Integration Tests
- Enrollment flow end-to-end
- Authentication flow end-to-end
- PAM integration
- Timeout handling
- Fallback behavior

### Security Tests
- Photo attack (print, screen)
- Video attack (replay)
- 3D mask attack (if applicable)
- Embedding theft attempt
- Brute force embedding comparison

### Performance Tests
- Frame capture latency
- Recognition speed
- Memory usage
- CPU usage during auth
- Multiple concurrent authentications

---

## Future Enhancements

- **GPU Acceleration:** Use CUDA for faster inference
- **Continuous Learning:** Update embeddings over time as face changes
- **Multi-user Support:** Recognize any enrolled user automatically
- **Remote Unlock:** Unlock via phone as backup
- **Audit Logging:** Detailed logs of auth attempts for security
- **Facial Expressions:** Use expressions as additional auth factor
- **Age Estimation:** Detect if user appears significantly different
- **Integration:** Sudo, SSH, application-level authentication

---

This architecture provides a solid foundation for a secure, performant, and user-friendly face authentication system for Linux!
