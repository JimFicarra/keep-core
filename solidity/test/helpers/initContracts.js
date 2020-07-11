const {contract, accounts} = require("@openzeppelin/test-environment")

const BLS = contract.fromArtifact('BLS');
const GroupSelection = contract.fromArtifact('GroupSelection');
const Groups = contract.fromArtifact('Groups');
const DKGResultVerification = contract.fromArtifact("DKGResultVerification");
const DelayFactor = contract.fromArtifact("DelayFactor");
const Reimbursements = contract.fromArtifact("Reimbursements");
const KeepRegistry = contract.fromArtifact("KeepRegistry");
const KeepToken = contract.fromArtifact('KeepToken');
const TokenGrant = contract.fromArtifact('TokenGrant');
const MinimumStakeSchedule = contract.fromArtifact('MinimumStakeSchedule');
const GrantStaking = contract.fromArtifact('GrantStaking');
const Locks = contract.fromArtifact('Locks');

async function initTokenStaking(
  tokenAddress,
  tokenGrantAddress,
  keepRegistryAddress,
  stakeInitializationPeriod,
  stakeUndelegationPeriod,
  TokenStakingEscrow,
  TokenStaking
) {
  let tokenStakingEscrow = await TokenStakingEscrow.new(
    tokenAddress,
    tokenGrantAddress,
    {from: accounts[0]}
  )

  await TokenStaking.detectNetwork()
  await TokenStaking.link(
    'MinimumStakeSchedule', 
    (await MinimumStakeSchedule.new({from: accounts[0]})).address
  )
  await TokenStaking.link(
    'GrantStaking',
    (await GrantStaking.new({from: accounts[0]})).address
  )
  await TokenStaking.link(
    'Locks',
    (await Locks.new({from: accounts[0]})).address
  )

  let tokenStaking = await TokenStaking.new(
    tokenAddress,
    tokenGrantAddress,
    tokenStakingEscrow.address,
    keepRegistryAddress,
    stakeInitializationPeriod,
    stakeUndelegationPeriod,
    {from: accounts[0]}
  );
  await tokenStakingEscrow.transferOwnership(
    tokenStaking.address,
    {from: accounts[0]}
  );

  return {
    tokenStakingEscrow: tokenStakingEscrow,
    tokenStaking: tokenStaking
  };
}

async function initContracts(TokenStaking, KeepRandomBeaconService,
  KeepRandomBeaconServiceImplV1, KeepRandomBeaconOperator) {

  let token, registry, stakingContract,
    serviceContractImplV1, serviceContractProxy, serviceContract,
    operatorContract;

  let dkgContributionMargin = 5, // 5% Represents DKG frequency of 1/20 (Every 20 entries trigger group selection)
    stakeInitializationPeriod = 30, // In seconds
    stakeUndelegationPeriod = 300; // In seconds

  token = await KeepToken.new({from: accounts[0]});
  tokenGrant = await TokenGrant.new(token.address, {from: accounts[0]});
  registry = await KeepRegistry.new({from: accounts[0]});

  // Initialize staking contract
  const stakingContracts = await initTokenStaking(
    token.address,
    tokenGrant.address,
    registry.address,
    stakeInitializationPeriod,
    stakeUndelegationPeriod,
    contract.fromArtifact('TokenStakingEscrow'),
    TokenStaking
  )
  stakingContract = stakingContracts.tokenStaking

  // Initialize Keep Random Beacon service contract
  serviceContractImplV1 = await KeepRandomBeaconServiceImplV1.new({from: accounts[0]});

  const initialize = serviceContractImplV1.contract.methods
      .initialize(
          dkgContributionMargin,
          registry.address,
      ).encodeABI();

  serviceContractProxy = await KeepRandomBeaconService.new(serviceContractImplV1.address, initialize, {from: accounts[0]});

  serviceContract = await KeepRandomBeaconServiceImplV1.at(serviceContractProxy.address);
  // Initialize Keep Random Beacon operator contract
  const bls = await BLS.new({from: accounts[0]});
  await KeepRandomBeaconOperator.detectNetwork()
  await KeepRandomBeaconOperator.link("BLS", bls.address);
  const groupSelection = await GroupSelection.new({from: accounts[0]});
  await Groups.detectNetwork()
  await Groups.link("BLS", bls.address);
  const groups = await Groups.new({from: accounts[0]});
  const delayFactor = await DelayFactor.new({from: accounts[0]});
  const dkgResultVerification = await DKGResultVerification.new({from: accounts[0]});

  const reimbursements = await Reimbursements.new({from: accounts[0]});

  await KeepRandomBeaconOperator.link("DelayFactor", delayFactor.address);
  await KeepRandomBeaconOperator.link("GroupSelection", groupSelection.address);
  await KeepRandomBeaconOperator.link("Groups", groups.address);
  await KeepRandomBeaconOperator.link("DKGResultVerification", dkgResultVerification.address);
  await KeepRandomBeaconOperator.link("Reimbursements", reimbursements.address);
  operatorContract = await KeepRandomBeaconOperator.new(
    serviceContractProxy.address,
    stakingContract.address,
    registry.address,
    {from: accounts[0]}
  );

  await registry.approveOperatorContract(operatorContract.address, {from: accounts[0]});

  // Set service contract owner as operator contract upgrader by default
  const operatorContractUpgrader = await serviceContractProxy.admin({from: accounts[0]})
  await registry.setOperatorContractUpgrader(serviceContract.address, operatorContractUpgrader, {from: accounts[0]});

  await serviceContract.addOperatorContract(operatorContract.address, {from: accounts[0]});

  let dkgGasEstimate = await operatorContract.dkgGasEstimate({from: accounts[0]});

  // Genesis should include payment to cover DKG cost to create first group
  let gasPriceCeiling = await operatorContract.gasPriceCeiling({from: accounts[0]});
  await operatorContract.genesis({value: dkgGasEstimate.mul(gasPriceCeiling), from: accounts[0]});

  return {
    registry: registry,
    token: token,
    stakingContract: stakingContract,
    serviceContract: serviceContract,
    operatorContract: operatorContract
  };
};

module.exports.initTokenStaking = initTokenStaking;
module.exports.initContracts = initContracts;