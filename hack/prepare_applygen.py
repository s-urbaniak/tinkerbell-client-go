"""Create a temporary, group/version ordered copy of the pinned API sources."""

import re
import shutil
import subprocess
from pathlib import Path

repo = Path(__file__).resolve().parents[1]
module = "github.com/s-urbaniak/tinkerbell-client-go"
api_source = Path(
    subprocess.check_output(
        ["go", "list", "-m", "-f", "{{.Dir}}", "github.com/tinkerbell/tinkerbell/api"],
        cwd=repo, text=True,
    ).strip()
)
output = repo / "internal/codegen/applyinput"
shutil.rmtree(output, ignore_errors=True)
roots = {
    "tinkerbell": {"Hardware", "Template", "Workflow", "WorkflowRuleSet"},
    "bmc": {"Job", "Machine", "Task"},
}
for group, kinds in roots.items():
    destination = output / group / "v1alpha1"
    destination.mkdir(parents=True)
    for source in (api_source / "v1alpha1" / group).glob("*.go"):
        if source.name.endswith("_test.go"):
            continue
        content = source.read_text()
        content = re.sub(r"(?m)^package \w+$", "package v1alpha1", content, count=1)
        content = content.replace(
            '"github.com/tinkerbell/tinkerbell/api/v1alpha1/bmc"',
            f'bmc "{module}/internal/codegen/applyinput/bmc/v1alpha1"',
        )
        if group == "bmc":
            content = re.sub(
                r'(metav1\.TypeMeta\s+`json:)""(`)',
                r'\1",inline"\2',
                content,
            )
        for kind in kinds:
            content = re.sub(
                rf"(?m)^(type {kind} struct \{{)",
                rf"// +genclient\n\1",
                content,
                count=1,
            )
        (destination / source.name).write_text(content)

    group_name = "tinkerbell.org" if group == "tinkerbell" else "bmc.tinkerbell.org"
    (destination / "doc.go").write_text(
        f"// +groupName={group_name}\n// Package v1alpha1 contains generator input.\npackage v1alpha1\n"
    )
