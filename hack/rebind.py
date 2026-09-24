"""Bind generated code to canonical Tinkerbell API types and resource names."""

from pathlib import Path
import re

root = Path(__file__).resolve().parents[1] / "generated"
module = "github.com/s-urbaniak/tinkerbell-client-go"
replacements = {
    f"{module}/internal/codegen/apis/bmc/v1alpha1":
        "github.com/tinkerbell/tinkerbell/api/v1alpha1/bmc",
    f"{module}/internal/codegen/apis/tinkerbell/v1alpha1":
        "github.com/tinkerbell/tinkerbell/api/v1alpha1/tinkerbell",
    f"{module}/internal/codegen/applyinput/bmc/v1alpha1":
        "github.com/tinkerbell/tinkerbell/api/v1alpha1/bmc",
    f"{module}/internal/codegen/applyinput/tinkerbell/v1alpha1":
        "github.com/tinkerbell/tinkerbell/api/v1alpha1/tinkerbell",
    "SchemeGroupVersion": "GroupVersion",
    'Resource: "hardwares"': 'Resource: "hardware"',
}

for path in root.rglob("*.go"):
    content = path.read_text()
    for old, new in replacements.items():
        content = content.replace(old, new)
    if "/listers/" in str(path):
        content = re.sub(
            r'([A-Za-z0-9]+)\.Resource\("([^"]+)"\)',
            r'\1.GroupVersion.WithResource("\2").GroupResource()',
            content,
        )
    path.write_text(content)
