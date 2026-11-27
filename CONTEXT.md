# SDC-Lite

YANG-schema-driven network configuration composition, inspection, diffing, attribution, and validation for network operating systems.

## Language

**Schema**:
The versioned YANG model set (including vendor models and deviations) that defines valid configuration structure, defaults, and constraints.
_Avoid_: Model, spec

**Intent**:
A named, prioritized configuration input layered onto a Workspace, including baseline device state such as `running`.
_Avoid_: Config file, config layer, overlay

**Workspace**:
The active combination of a Schema and zero or more Intents against which views, diffs, blame attributions, and validations are evaluated.
_Avoid_: Candidate, session, instance

**Target**:
A named, durable Workspace whose Schema reference and Intents survive across invocations.
_Avoid_: Profile, device, store

**Pipeline**:
An ordered sequence or live stream of Steps evaluated against a single Workspace.
_Avoid_: Batch, workflow, script

**Step**:
A single unit of work in a Pipeline that invokes one Method and yields a corresponding outcome.
_Avoid_: Command, instruction, call

**Method**:
The canonical operation name a Step invokes against a Workspace.
_Avoid_: Action, command type, subcommand
