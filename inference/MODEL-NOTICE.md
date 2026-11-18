# Original mark classifier

`bubble-marks-v1.safetensors` is an original development model produced by `train_bubbles.py`. It classifies a calibrated bubble patch as unfilled, filled or partial/ambiguous; it does not identify people, solve questions or award marks.

Training: 10,000 original synthetic patches, balanced across three classes, with randomized printed digits, circle geometry, ink/background color, blur, rotation and JPEG quality. A separate generated holdout of 2,000 patches produced 1,999 correct classes in the recorded run. Seed, epochs, elapsed time and limitations are in `bubble-marks-v1.training.json`. Training never reads institution/owner images. Rendered digit patches use Pillow's installed default font; dependency/font licenses remain their own.

The model weights and original synthetic fixture images may be reused under CC0-1.0, consistent with the demonstration-content grant in docs/PROVENANCE.md. No source-code license is assigned for the owner. Third-party LightOn/Qwen weights retain their own Apache-2.0 licenses.

This is a small synthetic training/evaluation set, not an independently validated production OMR dataset. Model margins are uncalibrated logit gaps, not probabilities of correctness. Real erasures, faint ink, checkmarks, photographs, diverse layouts and calibration mistakes need broader independent evaluation. The Go review pipeline preserves measured pixels, flags disagreement and never converts classifier output directly into academic marks.

Weights are a versioned artifact. Retraining may differ at the bit level across GPU/library versions. Deploy a new reviewed model revision rather than silently replacing a published artifact. Retain source image, calibration and actual weight digest with every diagnostic run.
