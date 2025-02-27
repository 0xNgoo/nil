{ lib
, stdenv
, biome
, python3
, callPackage
, npmHooks
, nodejs
, enableTesting ? false
, cypress
}:

stdenv.mkDerivation rec {
  name = "explorer";
  pname = "nilexplorer";
  src = lib.sourceByRegex ./.. [
    "package.json"
    "package-lock.json"
    "^niljs(/.*)?$"
    "^smart-contracts(/.*)?$"
    "biome.json"
    "^explorer_frontend(/.*)?$"
    "^explorer_backend(/.*)?$"
  ];

  npmDeps = (callPackage ./npmdeps.nix { });

  NODE_PATH = "$npmDeps";

  nativeBuildInputs = [
    nodejs
    npmHooks.npmConfigHook
    biome
    python3
  ];

  dontConfigure = true;

  preUnpack = ''
    echo "Setting UV_USE_IO_URING=0 to work around the io_uring kernel bug"
    export UV_USE_IO_URING=0

    export CYPRESS_INSTALL_BINARY=0
    export CYPRESS_RUN_BINARY=${cypress}/bin/Cypress
  '';

  buildPhase = ''
    patchShebangs explorer_frontend/node_modules

    (cd smart-contracts; npm ci && npm run build)
    (cd niljs; npm ci && npm run build)

    (cd explorer_frontend; npm ci && npm run build)
    (cd explorer_backend; npm ci && npm run build)
  '';

  doCheck = enableTesting;

  checkPhase = ''
    export BIOME_BINARY=${biome}/bin/biome

    echo "Checking explorer frontend"
    (cd explorer_frontend; npm run lint;)

    echo "Checking explorer backend"
    (cd explorer_backend; npm run lint;)

    echo "tests finished successfully"
  '';

  installPhase = ''
    mkdir -p $out
    mv explorer_frontend/ $out/frontend
    mv explorer_backend/ $out/backend
  '';
}
